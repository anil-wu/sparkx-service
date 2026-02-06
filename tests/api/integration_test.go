package api

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/anil-wu/spark-x/tests"
)

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
	Status      string `json:"status"` // active | archived
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
	FileFormat   string `json:"fileFormat"`
	SizeBytes    int64  `json:"sizeBytes"`
	Hash         string `json:"hash"`
	ContentType  string `json:"contentType"`
}

type PreUploadResp struct {
	UploadUrl     string `json:"uploadUrl"`
	FileId        int64  `json:"fileId"`
	VersionId     int64  `json:"versionId"`
	VersionNumber int64  `json:"versionNumber"`
	ContentType   string `json:"contentType"`
}

type ProjectFileItem struct {
	Id               int64  `json:"id"`
	ProjectId        int64  `json:"projectId"`
	Name             string `json:"name"`
	FileCategory     string `json:"fileCategory"`
	FileFormat       string `json:"fileFormat"`
	CurrentVersionId int64  `json:"currentVersionId"`
	VersionId        int64  `json:"versionId"`
	VersionNumber    int64  `json:"versionNumber"`
	SizeBytes        int64  `json:"sizeBytes"`
	Hash             string `json:"hash"`
	CreatedAt        string `json:"createdAt"`
	StorageKey       string `json:"storageKey"`
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
	StorageKey    string `json:"storageKey"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
	CreatedBy     int64  `json:"createdBy"`
}

type FileVersionListResp struct {
	List []FileVersionItem `json:"list"`
	Page PageResp          `json:"page"`
}

// BaseResp 基础响应
type BaseResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func TestWorkflow(t *testing.T) {
	if os.Getenv("SPARKX_INTEGRATION") == "" {
		t.Skip("SPARKX_INTEGRATION is not set")
	}

	rand.Seed(time.Now().UnixNano())
	client := &tests.TestClient{T: t}

	// ==========================================
	// 1. Auth & Users
	// ==========================================

	// Create User A
	emailA := tests.GetenvDefault("SPARKX_IT_EMAIL_A", "sparkx_it_user_a@example.com")
	password := tests.GetenvDefault("SPARKX_IT_PASSWORD", "password123")
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
	emailB := tests.GetenvDefault("SPARKX_IT_EMAIL_B", "sparkx_it_user_b@example.com")
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
			tests.DeleteProjectBestEffort(t, loginRespA.Token, projectResp.Id)
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
		FileFormat:   "txt",
		SizeBytes:    2048,
		Hash:         "testhash123456",
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
	client.Do("DELETE", fmt.Sprintf("/projects/%d", projectResp.Id), nil, &deleteProjectResp)
	if deleteProjectResp.Code != 0 {
		t.Fatalf("DeleteProject failed: code=%d msg=%s", deleteProjectResp.Code, deleteProjectResp.Msg)
	}
}
