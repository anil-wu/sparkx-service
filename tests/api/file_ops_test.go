package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
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
// File Operations Test (Download, Delete, Rollback)
// ==========================================

// TestFileOperations 测试文件下载、删除和版本回滚功能
func TestFileOperations(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &tests.TestClient{T: t}

	// ==========================================
	// 1. 创建测试用户和项目
	// ==========================================
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test_ops_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	t.Logf("Step 1: Create and login user (%s)", email)
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
	t.Logf("Login successful, userId: %d", loginResp.UserId)

	// 创建项目
	projectName := fmt.Sprintf("FileOpsTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginResp.UserId,
		Name:        projectName,
		Description: "Test project for file operations",
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
	// 2. 上传测试文件（创建多个版本）
	// ==========================================
	t.Log("========================================")
	t.Log("Step 2: Upload test file with multiple versions")
	t.Log("========================================")

	// 获取一个测试文件
	textFiles := data.GetTextFiles()
	if len(textFiles) == 0 {
		t.Fatal("No test files available")
	}
	testFile := textFiles[0]

	// 上传第一个版本
	fileId := uploadFileVersion(t, client, projectResp.Id, testFile.Name, testFile.Category, testFile.Format, testFile.Content)
	t.Logf("Created file with ID: %d", fileId)

	// 上传第二个版本（修改内容）
	version2Content := append(testFile.Content, []byte("\n// Version 2 content")...)
	uploadFileVersion(t, client, projectResp.Id, testFile.Name, testFile.Category, testFile.Format, version2Content)
	t.Logf("Uploaded version 2")

	// 上传第三个版本
	version3Content := append(testFile.Content, []byte("\n// Version 3 content")...)
	uploadFileVersion(t, client, projectResp.Id, testFile.Name, testFile.Category, testFile.Format, version3Content)
	t.Logf("Uploaded version 3")

	// ==========================================
	// 3. 测试文件下载接口
	// ==========================================
	t.Log("========================================")
	t.Log("Step 3: Test Download File API")
	t.Log("========================================")

	var downloadResp DownloadFileResp
	client.Get(fmt.Sprintf("/files/%d/download", fileId), &downloadResp)

	if downloadResp.DownloadUrl == "" {
		t.Fatal("Download URL is empty")
	}
	if downloadResp.ExpiresAt == "" {
		t.Fatal("ExpiresAt is empty")
	}
	t.Logf("Download URL generated successfully")
	t.Logf("  - URL length: %d", len(downloadResp.DownloadUrl))
	t.Logf("  - Expires at: %s", downloadResp.ExpiresAt)

	// 验证下载 URL 格式是否正确
	if !strings.HasPrefix(downloadResp.DownloadUrl, "http") {
		t.Fatal("Download URL should start with http")
	}
	if !strings.Contains(downloadResp.DownloadUrl, "?") {
		t.Fatal("Download URL should contain query parameters (signature)")
	}
	t.Logf("Download URL format is valid")

	// ==========================================
	// 4. 测试版本回滚接口
	// ==========================================
	t.Log("========================================")
	t.Log("Step 4: Test Rollback Version API")
	t.Log("========================================")

	// 获取当前版本列表
	var versionListResp FileVersionListResp
	client.Get(fmt.Sprintf("/files/%d/versions", fileId), &versionListResp)
	if len(versionListResp.List) == 0 {
		t.Fatal("No versions found")
	}
	t.Logf("File has %d versions", len(versionListResp.List))

	// 回滚到第一个版本
	rollbackReq := RollbackVersionReq{
		VersionNumber: 1,
	}
	var rollbackResp BaseResp
	client.Post(fmt.Sprintf("/files/%d/rollback", fileId), rollbackReq, &rollbackResp)
	if rollbackResp.Code != 0 {
		t.Fatalf("Rollback failed: code=%d msg=%s", rollbackResp.Code, rollbackResp.Msg)
	}
	t.Logf("Successfully rolled back to version 1")

	// 验证回滚结果
	var fileListResp ProjectFileListResp
	client.Get(fmt.Sprintf("/projects/%d/files", projectResp.Id), &fileListResp)

	var foundFile *ProjectFileItem
	for i := range fileListResp.List {
		if fileListResp.List[i].Id == fileId {
			foundFile = &fileListResp.List[i]
			break
		}
	}
	if foundFile == nil {
		t.Fatal("File not found after rollback")
	}

	// 验证当前版本是否为版本 1
	if foundFile.VersionNumber != 1 {
		t.Fatalf("Rollback failed: expected version 1, got version %d", foundFile.VersionNumber)
	}
	t.Logf("Verified: current version is now %d", foundFile.VersionNumber)

	// ==========================================
	// 5. 测试文件删除接口
	// ==========================================
	t.Log("========================================")
	t.Log("Step 5: Test Delete File API")
	t.Log("========================================")

	var deleteResp BaseResp
	client.Do("DELETE", fmt.Sprintf("/files/%d", fileId), nil, &deleteResp)
	if deleteResp.Code != 0 {
		t.Fatalf("Delete file failed: code=%d msg=%s", deleteResp.Code, deleteResp.Msg)
	}
	t.Logf("File deleted successfully")

	// 验证文件已被删除（软删除，文件列表中不应出现）
	client.Get(fmt.Sprintf("/projects/%d/files", projectResp.Id), &fileListResp)
	for _, f := range fileListResp.List {
		if f.Id == fileId {
			t.Fatal("Deleted file still appears in file list")
		}
	}
	t.Logf("Verified: deleted file no longer appears in list")

	// ==========================================
	// 6. 测试删除不存在的文件（错误处理）
	// ==========================================
	t.Log("========================================")
	t.Log("Step 6: Test error handling - delete non-existent file")
	t.Log("========================================")

	// 删除不存在的文件应该返回错误（不解析响应体）
	client.DoWithStatusCheck("DELETE", fmt.Sprintf("/files/%d", 999999), nil, nil, false)
	t.Log("Delete non-existent file returned expected error")

	// ==========================================
	// 7. 测试回滚到不存在的版本（错误处理）
	// ==========================================
	t.Log("========================================")
	t.Log("Step 7: Test error handling - rollback to non-existent version")
	t.Log("========================================")

	// 创建一个新文件用于测试
	newFileId := uploadFileVersion(t, client, projectResp.Id, "test_rollback_error.txt", "text", "txt", []byte("test content"))
	t.Logf("Created new file with ID: %d", newFileId)

	// 尝试回滚到不存在的版本 999（应该返回错误）
	invalidRollbackReq := RollbackVersionReq{
		VersionNumber: 999,
	}
	client.DoWithStatusCheck("POST", fmt.Sprintf("/files/%d/rollback", newFileId), invalidRollbackReq, nil, false)
	t.Log("Rolling back to non-existent version correctly returned error")

	// 清理
	client.DoWithStatusCheck("DELETE", fmt.Sprintf("/files/%d", newFileId), nil, nil, false)

	t.Log("========================================")
	t.Log("All file operations tests PASSED!")
	t.Log("========================================")
}

// TestGetFileContent 测试文件内容代理访问接口
func TestGetFileContent(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &tests.TestClient{T: t}

	// 创建用户和项目
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test_content_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	loginReq := LoginReq{
		LoginType: "email",
		Email:     email,
		Password:  password,
	}
	var loginResp LoginResp
	client.Post("/auth/login", loginReq, &loginResp)
	if loginResp.UserId == 0 {
		t.Fatal("Login failed")
	}
	client.SetToken(loginResp.Token)

	projectName := fmt.Sprintf("ContentTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginResp.UserId,
		Name:        projectName,
		Description: "Test file content access",
	}
	var projectResp ProjectResp
	client.Post("/projects", createProjReq, &projectResp)
	if projectResp.Id == 0 {
		t.Fatal("Create project failed")
	}

	t.Cleanup(func() {
		if projectResp.Id > 0 {
			tests.DeleteProjectBestEffort(t, loginResp.Token, projectResp.Id)
		}
	})

	// 上传测试文件
	textFiles := data.GetTextFiles()
	testFile := textFiles[0]
	fileId := uploadFileVersion(t, client, projectResp.Id, testFile.Name, testFile.Category, testFile.Format, testFile.Content)
	t.Logf("Created file with ID: %d", fileId)

	// 测试获取文件内容
	t.Log("Testing: Get file content via proxy")

	// 创建 HTTP 请求获取文件内容
	httpClient := &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	req, err := http.NewRequest("GET", tests.GetBaseURL()+fmt.Sprintf("/files/%d/content", fileId), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to get file content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get file content failed with status: %d", resp.StatusCode)
	}

	// 读取内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// 验证 Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		t.Log("Warning: Content-Type header is empty")
	} else {
		t.Logf("Content-Type: %s", contentType)
	}

	// 验证内容长度
	if len(content) == 0 {
		t.Fatal("File content is empty")
	}
	t.Logf("File content received: %d bytes", len(content))

	// 清理
	client.DoWithStatusCheck("DELETE", fmt.Sprintf("/files/%d", fileId), nil, nil, false)

	t.Log("Get file content test PASSED!")
}

// uploadFileVersion 辅助函数：上传文件版本
func uploadFileVersion(t *testing.T, client *tests.TestClient, projectId int64, fileName, category, format string, content []byte) int64 {
	sizeBytes := int64(len(content))
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// PreUpload
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
		t.Fatal("PreUpload failed")
	}

	// 上传文件到 OSS
	uploadReq, err := http.NewRequest("PUT", preUploadResp.UploadUrl, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Failed to create upload request: %v", err)
	}

	contentType := preUploadResp.ContentType
	if contentType == "" {
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

	return preUploadResp.FileId
}

// TestFileDownloadPermission 测试文件下载权限（项目成员才能下载）
func TestFileDownloadPermission(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())

	// 创建用户 A（项目所有者）
	clientA := &tests.TestClient{T: t}
	timestamp := time.Now().UnixNano()
	emailA := fmt.Sprintf("test_owner_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	loginReqA := LoginReq{
		LoginType: "email",
		Email:     emailA,
		Password:  password,
	}
	var loginRespA LoginResp
	clientA.Post("/auth/login", loginReqA, &loginRespA)
	if loginRespA.UserId == 0 {
		t.Fatal("Login A failed")
	}
	clientA.SetToken(loginRespA.Token)

	// 创建项目
	projectName := fmt.Sprintf("PermissionTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginRespA.UserId,
		Name:        projectName,
		Description: "Test project for permission",
	}
	var projectResp ProjectResp
	clientA.Post("/projects", createProjReq, &projectResp)
	if projectResp.Id == 0 {
		t.Fatal("Create project failed")
	}

	t.Cleanup(func() {
		if projectResp.Id > 0 {
			tests.DeleteProjectBestEffort(t, loginRespA.Token, projectResp.Id)
		}
	})

	// 上传文件
	textFiles := data.GetTextFiles()
	fileId := uploadFileVersion(t, clientA, projectResp.Id, textFiles[0].Name, textFiles[0].Category, textFiles[0].Format, textFiles[0].Content)

	// 用户 A 应该能下载
	t.Log("Testing: Owner can download file")
	var downloadResp DownloadFileResp
	clientA.Get(fmt.Sprintf("/files/%d/download", fileId), &downloadResp)
	if downloadResp.DownloadUrl == "" {
		t.Fatal("Owner should be able to download file")
	}
	t.Log("Owner download: PASSED")

	// 创建用户 B（非项目成员）
	clientB := &tests.TestClient{T: t}
	emailB := fmt.Sprintf("test_outsider_%d@example.com", timestamp+1)
	loginReqB := LoginReq{
		LoginType: "email",
		Email:     emailB,
		Password:  password,
	}
	var loginRespB LoginResp
	clientB.Post("/auth/login", loginReqB, &loginRespB)
	if loginRespB.UserId == 0 {
		t.Fatal("Login B failed")
	}
	clientB.SetToken(loginRespB.Token)

	// 用户 B 不应该能下载（应该返回权限错误）
	t.Log("Testing: Non-member cannot download file")
	clientB.DoWithStatusCheck("GET", fmt.Sprintf("/files/%d/download", fileId), nil, nil, false)
	t.Log("Non-member download correctly rejected")
}

// TestFileDeletePermission 测试文件删除权限（只有 owner/admin 能删除）
func TestFileDeletePermission(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())

	// 创建用户 A（项目所有者）
	clientA := &tests.TestClient{T: t}
	timestamp := time.Now().UnixNano()
	emailA := fmt.Sprintf("test_del_owner_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	loginReqA := LoginReq{
		LoginType: "email",
		Email:     emailA,
		Password:  password,
	}
	var loginRespA LoginResp
	clientA.Post("/auth/login", loginReqA, &loginRespA)
	if loginRespA.UserId == 0 {
		t.Fatal("Login A failed")
	}
	clientA.SetToken(loginRespA.Token)

	// 创建项目
	projectName := fmt.Sprintf("DeletePermTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginRespA.UserId,
		Name:        projectName,
		Description: "Test delete permission",
	}
	var projectResp ProjectResp
	clientA.Post("/projects", createProjReq, &projectResp)
	if projectResp.Id == 0 {
		t.Fatal("Create project failed")
	}

	t.Cleanup(func() {
		if projectResp.Id > 0 {
			tests.DeleteProjectBestEffort(t, loginRespA.Token, projectResp.Id)
		}
	})

	// 上传文件
	textFiles := data.GetTextFiles()
	fileId := uploadFileVersion(t, clientA, projectResp.Id, textFiles[0].Name, textFiles[0].Category, textFiles[0].Format, textFiles[0].Content)

	// 创建用户 B（developer 角色）
	clientB := &tests.TestClient{T: t}
	emailB := fmt.Sprintf("test_dev_%d@example.com", timestamp+1)
	loginReqB := LoginReq{
		LoginType: "email",
		Email:     emailB,
		Password:  password,
	}
	var loginRespB LoginResp
	clientB.Post("/auth/login", loginReqB, &loginRespB)
	if loginRespB.UserId == 0 {
		t.Fatal("Login B failed")
	}
	clientB.SetToken(loginRespB.Token)

	// 邀请用户 B 作为 developer
	inviteReq := InviteMemberReq{
		UserId:        loginRespA.UserId,
		InvitedUserId: loginRespB.UserId,
		ProjectId:     projectResp.Id,
		Role:          "developer",
	}
	var inviteResp BaseResp
	clientA.Post(fmt.Sprintf("/projects/%d/invite", projectResp.Id), inviteReq, &inviteResp)
	if inviteResp.Code != 0 {
		t.Fatalf("Invite member failed: %s", inviteResp.Msg)
	}

	// 用户 B（developer）尝试删除文件应该失败
	t.Log("Testing: Developer cannot delete file")
	clientB.DoWithStatusCheck("DELETE", fmt.Sprintf("/files/%d", fileId), nil, nil, false)
	t.Log("Developer delete correctly rejected")

	// 用户 A（owner）删除文件应该成功
	t.Log("Testing: Owner can delete file")
	var deleteRespA BaseResp
	clientA.Do("DELETE", fmt.Sprintf("/files/%d", fileId), nil, &deleteRespA)
	if deleteRespA.Code != 0 {
		t.Fatalf("Owner should be able to delete file: %s", deleteRespA.Msg)
	}
	t.Log("Owner delete: PASSED")
}

// TestRollbackCurrentVersion 测试回滚到当前版本（应该返回错误）
func TestRollbackCurrentVersion(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &tests.TestClient{T: t}

	// 创建用户和项目
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test_rollback_curr_%d@example.com", timestamp)
	password := fmt.Sprintf("Pass%d!", rand.Intn(1000000))

	loginReq := LoginReq{
		LoginType: "email",
		Email:     email,
		Password:  password,
	}
	var loginResp LoginResp
	client.Post("/auth/login", loginReq, &loginResp)
	if loginResp.UserId == 0 {
		t.Fatal("Login failed")
	}
	client.SetToken(loginResp.Token)

	projectName := fmt.Sprintf("RollbackCurrTest_%d", rand.Intn(10000))
	createProjReq := CreateProjectReq{
		UserId:      loginResp.UserId,
		Name:        projectName,
		Description: "Test rollback to current version",
	}
	var projectResp ProjectResp
	client.Post("/projects", createProjReq, &projectResp)
	if projectResp.Id == 0 {
		t.Fatal("Create project failed")
	}

	t.Cleanup(func() {
		if projectResp.Id > 0 {
			tests.DeleteProjectBestEffort(t, loginResp.Token, projectResp.Id)
		}
	})

	// 上传文件
	textFiles := data.GetTextFiles()
	fileId := uploadFileVersion(t, client, projectResp.Id, textFiles[0].Name, textFiles[0].Category, textFiles[0].Format, textFiles[0].Content)

	// 获取当前版本号
	var fileListResp ProjectFileListResp
	client.Get(fmt.Sprintf("/projects/%d/files", projectResp.Id), &fileListResp)

	var currentVersion int64
	for _, f := range fileListResp.List {
		if f.Id == fileId {
			currentVersion = f.VersionNumber
			break
		}
	}

	t.Logf("Current version: %d", currentVersion)

	// 尝试回滚到当前版本（应该返回错误）
	rollbackReq := RollbackVersionReq{
		VersionNumber: currentVersion,
	}
	client.DoWithStatusCheck("POST", fmt.Sprintf("/files/%d/rollback", fileId), rollbackReq, nil, false)
	t.Log("Correctly rejected rollback to current version")

	// 清理
	client.Do("DELETE", fmt.Sprintf("/files/%d", fileId), nil, &BaseResp{})
}
