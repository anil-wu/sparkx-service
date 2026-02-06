package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anil-wu/spark-x/tests"
	"github.com/anil-wu/spark-x/tests/api/data"
)

// ==========================================
// File Upload Complete Test
// ==========================================

// TestFileUploadComplete 测试完整的文件上传流程，包括直传 OSS
// 使用随机生成的账号，自动创建用户进行测试
func TestFileUploadComplete(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &tests.TestClient{T: t}

	// ==========================================
	// 1. 自动生成账号并登录（自动创建用户）
	// ==========================================
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	t.Logf("Step 1: Auto-create and login user (%s)", email)
	loginReq := LoginReq{
		LoginType: "email",
		Email:     email,
		Password:  password,
	}
	var loginResp LoginResp
	client.Post("/auth/login", loginReq, &loginResp)
	if loginResp.UserId == 0 {
		t.Fatal("Login/Register failed")
	}
	client.SetToken(loginResp.Token)
	t.Logf("Login successful, userId: %d, isNewUser: %v", loginResp.UserId, loginResp.Created)

	// 创建项目
	projectName := fmt.Sprintf("FileUploadTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginResp.UserId,
		Name:        projectName,
		Description: "Test project for file upload",
	}
	var projectResp ProjectResp
	client.Post("/projects", createProjReq, &projectResp)
	if projectResp.Id == 0 {
		t.Fatal("Create project failed")
	}
	t.Logf("Project created, projectId: %d", projectResp.Id)

	t.Cleanup(func() {
		if projectResp.Id > 0 {
			tests.DeleteProjectBestEffort(t, loginResp.Token, projectResp.Id)
		}
	})

	// ==========================================
	// 2. 测试上传所有文本类型文件
	// ==========================================
	textFiles := data.GetTextFiles()
	for i, tf := range textFiles {
		t.Log("========================================")
		t.Logf("Test %d: Upload %s file", i+1, tf.Format)
		t.Log("========================================")
		testUploadFile(t, client, projectResp.Id, tf.Name, tf.Category, tf.Format, tf.Content)
	}

	// ==========================================
	// 3. 测试上传所有图片类型文件
	// ==========================================
	imageFiles := data.GetImageFiles()
	for i, tf := range imageFiles {
		t.Log("========================================")
		t.Logf("Test %d: Upload %s image", i+len(textFiles)+1, tf.Format)
		t.Log("========================================")
		testUploadFile(t, client, projectResp.Id, tf.Name, tf.Category, tf.Format, tf.Content)
	}

	t.Log("========================================")
	t.Logf("All %d file upload tests PASSED!", len(textFiles)+len(imageFiles))
	t.Log("========================================")
}

// testUploadFile 通用的文件上传测试函数
func testUploadFile(t *testing.T, client *tests.TestClient, projectId int64, fileName, category, format string, content []byte) {
	sizeBytes := int64(len(content))
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	t.Logf("Preparing test file:")
	t.Logf("  - File name: %s", fileName)
	t.Logf("  - Category: %s", category)
	t.Logf("  - Format: %s", format)
	t.Logf("  - Size: %d bytes", sizeBytes)
	t.Logf("  - SHA256: %s", hashStr)

	// 调用 PreUpload API
	t.Log("Calling PreUpload API...")
	preUploadReq := PreUploadReq{
		ProjectId:    projectId,
		Name:         fileName,
		FileCategory: category,
		FileFormat:   format,
		SizeBytes:    sizeBytes,
		Hash:         hashStr,
	}
	var preUploadResp PreUploadResp
	client.Post("/files/preupload", preUploadReq, &preUploadResp)

	if preUploadResp.FileId == 0 {
		t.Fatal("PreUpload failed: FileId is 0")
	}
	if preUploadResp.UploadUrl == "" {
		t.Fatal("PreUpload failed: UploadUrl is empty")
	}
	if preUploadResp.VersionId == 0 {
		t.Fatal("PreUpload failed: VersionId is 0")
	}

	t.Logf("PreUpload successful:")
	t.Logf("  - FileId: %d", preUploadResp.FileId)
	t.Logf("  - VersionId: %d", preUploadResp.VersionId)
	t.Logf("  - VersionNumber: %d", preUploadResp.VersionNumber)
	t.Logf("  - ContentType: %s", preUploadResp.ContentType)

	// 解析并打印 OSS 路径
	uploadUrl := preUploadResp.UploadUrl
	if idx := strings.Index(uploadUrl, "?"); idx > 0 {
		objectPath := uploadUrl[strings.LastIndex(uploadUrl[:idx], "/")+1 : idx]
		bucketHost := uploadUrl[:strings.Index(uploadUrl, ".")]
		t.Logf("  - OSS Bucket: %s", bucketHost)
		t.Logf("  - OSS Object Path: %s", objectPath)
	}

	// 直传文件到 OSS
	t.Log("Uploading file to OSS directly...")
	uploadReq, err := http.NewRequest("PUT", preUploadResp.UploadUrl, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to create upload request: %v", err)
	}
	// 使用服务端返回的 Content-Type，确保与签名一致
	contentType := preUploadResp.ContentType
	if contentType == "" {
		// 回退到根据类别推断
		if category == "image" {
			contentType = "image/png"
		} else {
			contentType = "text/plain"
		}
	}
	uploadReq.Header.Set("Content-Type", contentType)

	uploadClient := &http.Client{Timeout: 30 * time.Second}
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		t.Fatalf("Failed to upload file to OSS: %v", err)
	}
	defer uploadResp.Body.Close()

	uploadRespBody, _ := io.ReadAll(uploadResp.Body)
	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("OSS upload failed. Status: %d, Body: %s", uploadResp.StatusCode, string(uploadRespBody))
	}
	t.Logf("OSS upload successful. Status: %d", uploadResp.StatusCode)

	// 验证文件列表
	t.Log("Verifying file list...")
	var fileListResp ProjectFileListResp
	client.Get(fmt.Sprintf("/projects/%d/files", projectId), &fileListResp)

	var foundFile *ProjectFileItem
	for i := range fileListResp.List {
		if fileListResp.List[i].Id == preUploadResp.FileId {
			foundFile = &fileListResp.List[i]
			break
		}
	}

	if foundFile == nil {
		t.Fatalf("Uploaded file not found in project file list. Total files: %d", len(fileListResp.List))
	}

	if foundFile.Name != fileName {
		t.Errorf("File name mismatch. Got %s, want %s", foundFile.Name, fileName)
	}
	if foundFile.SizeBytes != sizeBytes {
		t.Errorf("File size mismatch. Got %d, want %d", foundFile.SizeBytes, sizeBytes)
	}
	if foundFile.Hash != hashStr {
		t.Errorf("File hash mismatch. Got %s, want %s", foundFile.Hash, hashStr)
	}

	t.Logf("File verification passed:")
	t.Logf("  - FileId: %d", foundFile.Id)
	t.Logf("  - Name: %s", foundFile.Name)
	t.Logf("  - Size: %d bytes", foundFile.SizeBytes)
	t.Logf("  - VersionNumber: %d", foundFile.VersionNumber)

	// 验证版本列表
	t.Log("Verifying file versions...")
	var versionListResp FileVersionListResp
	client.Get(fmt.Sprintf("/files/%d/versions", preUploadResp.FileId), &versionListResp)

	if len(versionListResp.List) == 0 {
		t.Fatal("No versions found for file")
	}

	latestVersion := versionListResp.List[0]
	if latestVersion.VersionNumber != preUploadResp.VersionNumber {
		t.Errorf("Version number mismatch. Got %d, want %d", latestVersion.VersionNumber, preUploadResp.VersionNumber)
	}
	if latestVersion.SizeBytes != sizeBytes {
		t.Errorf("Version size mismatch. Got %d, want %d", latestVersion.SizeBytes, sizeBytes)
	}
	if latestVersion.Hash != hashStr {
		t.Errorf("Version hash mismatch. Got %s, want %s", latestVersion.Hash, hashStr)
	}

	t.Logf("Version verification passed:")
	t.Logf("  - VersionId: %d", latestVersion.Id)
	t.Logf("  - VersionNumber: %d", latestVersion.VersionNumber)
	t.Logf("  - CreatedBy: %d", latestVersion.CreatedBy)
}
