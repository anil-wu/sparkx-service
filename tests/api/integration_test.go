package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	defaultBaseURL = "https://localhost:8890/api/v1"
)

type Response struct {
	Code int32           `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"` // Delay parsing
}

type BaseResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

// Data structs matching API
type LoginReq struct {
	LoginType string `json:"loginType"` // email | google
	Email     string `json:"email,optional"`
	Password  string `json:"password,optional"`
	IdToken   string `json:"idToken,optional"`
}

type LoginResp struct {
	UserId  int64  `json:"userId"`
	Created bool   `json:"created"`
	Token   string `json:"token"`
}

type CreateProjectReq struct {
	UserId      int64  `json:"userId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectResp struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerId     int64  `json:"ownerId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type UpdateProjectReq struct {
	Id          int64  `json:"id"` // path param in logic, but here helper handles path
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type PageResp struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
	Total    int64 `json:"total"`
}

type ProjectListResp struct {
	List []ProjectResp `json:"list"`
	Page PageResp      `json:"page"`
}

// Users
type UserInfoResp struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type UserListResp struct {
	List []UserInfoResp `json:"list"`
	Page PageResp       `json:"page"`
}

type UpdateUserReq struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
}

// Projects
type InviteMemberReq struct {
	UserId        int64  `json:"userId"`
	InvitedUserId int64  `json:"invitedUserId"`
	ProjectId     int64  `json:"projectId"` // path param in logic
	Role          string `json:"role"`
}

// Files
type PreUploadReq struct {
	ProjectId    int64  `json:"projectId"`
	Name         string `json:"name"`
	FileCategory string `json:"fileCategory"`
	SizeBytes    int64  `json:"sizeBytes"`
	Hash         string `json:"hash"`
	MimeType     string `json:"mimeType"`
}

type PreUploadResp struct {
	UploadUrl     string `json:"uploadUrl"`
	FileId        int64  `json:"fileId"`
	VersionId     int64  `json:"versionId"`
	VersionNumber int64  `json:"versionNumber"`
}

type ProjectFileItem struct {
	Id            int64  `json:"id"`
	ProjectId     int64  `json:"projectId"`
	Name          string `json:"name"`
	FileCategory  string `json:"fileCategory"`
	VersionId     int64  `json:"versionId"`
	VersionNumber int64  `json:"versionNumber"`
	SizeBytes     int64  `json:"sizeBytes"`
	Hash          string `json:"hash"`
	MimeType      string `json:"mimeType"`
	CreatedAt     string `json:"createdAt"`
	StoragePath   string `json:"storagePath"`
}

type ProjectFileListResp struct {
	List []ProjectFileItem `json:"list"`
	Page PageResp          `json:"page"`
}

type FileVersionItem struct {
	Id            int64  `json:"id"`
	FileId        int64  `json:"fileId"`
	VersionNumber int64  `json:"versionNumber"`
	SizeBytes     int64  `json:"sizeBytes"`
	Hash          string `json:"hash"`
	StoragePath   string `json:"storagePath"`
	MimeType      string `json:"mimeType"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	CreatedBy     int64  `json:"createdBy"`
}

type FileVersionListResp struct {
	List []FileVersionItem `json:"list"`
	Page PageResp          `json:"page"`
}

// Helper client
type Client struct {
	t     *testing.T
	token string
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getBaseURL() string {
	v := os.Getenv("SPARKX_BASE_URL")
	if v == "" {
		return defaultBaseURL
	}
	return strings.TrimRight(v, "/")
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) Post(path string, body interface{}, target interface{}) {
	c.do("POST", path, body, target)
}

func (c *Client) Put(path string, body interface{}, target interface{}) {
	c.do("PUT", path, body, target)
}

func (c *Client) Get(path string, target interface{}) {
	c.do("GET", path, nil, target)
}

func (c *Client) Delete(path string) {
	c.do("DELETE", path, nil, nil)
}

func (c *Client) do(method, path string, body interface{}, target interface{}) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("Failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, getBaseURL()+path, bodyReader)
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add JWT token if available
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Skip TLS verification for self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		c.t.Fatalf("API Error %d: %s", resp.StatusCode, string(respBytes))
	}

	if target != nil {
		if err := json.Unmarshal(respBytes, target); err != nil {
			c.t.Fatalf("Failed to unmarshal response: %v. Body: %s", err, string(respBytes))
		}
	}
}

func deleteProjectBestEffort(t *testing.T, token string, id int64) {
	req, err := http.NewRequest("DELETE", getBaseURL()+fmt.Sprintf("/projects/%d", id), nil)
	if err != nil {
		t.Logf("Cleanup failed to create request: %v", err)
		return
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// Skip TLS verification for self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Cleanup request failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		t.Logf("Cleanup API Error %d: %s", resp.StatusCode, string(b))
	}
}

func TestWorkflow(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &Client{t: t}

	// ==========================================
	// 1. Auth & Users
	// ==========================================

	// Create User A
	emailA := getenvDefault("SPARKX_IT_EMAIL_A", "sparkx_it_user_a@example.com")
	password := getenvDefault("SPARKX_IT_PASSWORD", "password123")
	t.Logf("Step 1.1: Register User A (%s)", emailA)
	loginReqA := LoginReq{
		LoginType: "email",
		Email:     emailA,
		Password:  password,
	}
	var loginRespA LoginResp
	client.Post("/auth/login", loginReqA, &loginRespA)
	if loginRespA.UserId == 0 {
		t.Fatal("Login A failed")
	}
	client.SetToken(loginRespA.Token)
	t.Logf("User A logged in, token received")

	// Create User B (for invitation)
	emailB := getenvDefault("SPARKX_IT_EMAIL_B", "sparkx_it_user_b@example.com")
	t.Logf("Step 1.2: Register User B (%s)", emailB)
	loginReqB := LoginReq{
		LoginType: "email",
		Email:     emailB,
		Password:  password,
	}
	var loginRespB LoginResp
	client.Post("/auth/login", loginReqB, &loginRespB)

	// Update User A
	t.Log("Step 1.3: Update User A")
	newUsername := "sparkx_it_user_a"
	updateUserReq := UpdateUserReq{Username: newUsername}
	var updateUserResp BaseResp
	client.Put(fmt.Sprintf("/users/%d", loginRespA.UserId), updateUserReq, &updateUserResp)
	if updateUserResp.Code != 0 {
		t.Fatalf("UpdateUser failed: code=%d msg=%s", updateUserResp.Code, updateUserResp.Msg)
	}

	// Get User A by Email
	t.Log("Step 1.4: Get User A by Email")
	var userResp UserInfoResp
	client.Get(fmt.Sprintf("/users/email/%s", emailA), &userResp)
	if userResp.Username != newUsername {
		t.Errorf("User update failed. Got %s, want %s", userResp.Username, newUsername)
	}

	// List Users
	t.Log("Step 1.5: List Users")
	var userListResp UserListResp
	client.Get("/users", &userListResp)
	if len(userListResp.List) == 0 {
		t.Log("Warning: User list is empty")
	}

	// ==========================================
	// 2. Projects
	// ==========================================

	// Create Project
	t.Log("Step 2.1: Create Project")
	projectName := fmt.Sprintf("Test Project %d", rand.Intn(1000))
	createProjReq := CreateProjectReq{
		UserId:      loginRespA.UserId,
		Name:        projectName,
		Description: "Integration Test Project",
	}
	var projectResp ProjectResp
	client.Post("/projects", createProjReq, &projectResp)
	t.Logf("Project created with ID: %d", projectResp.Id)
	t.Cleanup(func() {
		if projectResp.Id > 0 {
			deleteProjectBestEffort(t, loginRespA.Token, projectResp.Id)
		}
	})

	// Get Project
	t.Log("Step 2.2: Get Project")
	var gotProject ProjectResp
	client.Get(fmt.Sprintf("/projects/%d", projectResp.Id), &gotProject)
	if gotProject.Id != projectResp.Id {
		t.Fatalf("GetProject mismatch. Got %d, want %d", gotProject.Id, projectResp.Id)
	}

	// Update Project
	t.Log("Step 2.3: Update Project")
	updateProjReq := UpdateProjectReq{
		Name:        projectName + " Updated",
		Description: "Integration Test Project Updated",
		Status:      "archived",
	}
	var updateProjectResp BaseResp
	client.Put(fmt.Sprintf("/projects/%d", projectResp.Id), updateProjReq, &updateProjectResp)
	if updateProjectResp.Code != 0 {
		t.Fatalf("UpdateProject failed: code=%d msg=%s", updateProjectResp.Code, updateProjectResp.Msg)
	}

	// List Projects
	t.Log("Step 2.4: List Projects")
	var projectListResp ProjectListResp
	client.Get("/projects", &projectListResp)
	foundProject := false
	for _, p := range projectListResp.List {
		if p.Id == projectResp.Id {
			foundProject = true
			break
		}
	}
	if !foundProject {
		t.Logf("Warning: project %d not found in list", projectResp.Id)
	}

	// Invite User B
	t.Log("Step 2.5: Invite User B to Project")
	inviteReq := InviteMemberReq{
		UserId:        loginRespA.UserId,
		InvitedUserId: loginRespB.UserId,
		Role:          "developer",
	}
	var inviteResp BaseResp
	client.Post(fmt.Sprintf("/projects/%d/invite", projectResp.Id), inviteReq, &inviteResp)
	if inviteResp.Code != 0 {
		t.Fatalf("InviteMember failed: code=%d msg=%s", inviteResp.Code, inviteResp.Msg)
	}

	// ==========================================
	// 3. Files
	// ==========================================

	// PreUpload File
	t.Log("Step 3.1: PreUpload File")
	preUploadReq := PreUploadReq{
		ProjectId:    projectResp.Id,
		Name:         "test_doc.txt",
		FileCategory: "text",
		SizeBytes:    2048,
		Hash:         "testhash123456",
		MimeType:     "text/plain",
	}
	var preUploadResp PreUploadResp
	client.Post("/files/preupload", preUploadReq, &preUploadResp)
	t.Logf("PreUpload successful. FileID: %d", preUploadResp.FileId)

	// List Project Files
	t.Log("Step 3.2: List Project Files")
	var fileListResp ProjectFileListResp
	client.Get(fmt.Sprintf("/projects/%d/files", projectResp.Id), &fileListResp)

	fileFound := false
	for _, f := range fileListResp.List {
		if f.Id == preUploadResp.FileId {
			fileFound = true
			break
		}
	}
	if !fileFound {
		t.Error("Uploaded file not found in project file list")
	}

	// List File Versions
	t.Log("Step 3.3: List File Versions")
	var versionListResp FileVersionListResp
	client.Get(fmt.Sprintf("/files/%d/versions", preUploadResp.FileId), &versionListResp)
	if len(versionListResp.List) == 0 {
		t.Error("No versions found for file")
	}

	// ==========================================
	// 4. Cleanup
	// ==========================================
	t.Log("Step 4: Cleanup - Delete Project")
	var deleteProjectResp BaseResp
	client.do("DELETE", fmt.Sprintf("/projects/%d", projectResp.Id), nil, &deleteProjectResp)
	if deleteProjectResp.Code != 0 {
		t.Fatalf("DeleteProject failed: code=%d msg=%s", deleteProjectResp.Code, deleteProjectResp.Msg)
	}
}
