# Spark-X API 列表

本文档列出了 `service/sparkx.api` 中定义的所有 API 接口信息。

## 基础信息
- **版本**: v1
- **前缀**: `/api/v1`

## 认证 (Auth)

### 登录/注册
- **接口**: `Login`
- **方法**: `POST`
- **路径**: `/auth/login`
- **请求**: `LoginReq`
  - `loginType` (string): 登录类型 - `email` | `google`
  - `email` (string, optional): 用户邮箱 (email 登录时需要)
  - `password` (string, optional): 用户密码 (email 登录时需要)
  - `idToken` (string, optional): Google ID Token (google 登录时需要)
- **响应**: `LoginResp`
  - `userId` (int64): 用户 ID
  - `created` (bool): 是否为新创建用户 (注册逻辑)
  - `token` (string): JWT Token，用于后续请求的认证

---

## 用户 (Users)

### 根据邮箱获取用户
- **接口**: `GetUserByEmail`
- **方法**: `GET`
- **路径**: `/users/email/:email`
- **请求**: `GetUserByEmailReq`
  - `email` (string, path): 用户邮箱
- **响应**: `UserInfoResp`
  - `id` (int64)
  - `username` (string)
  - `email` (string)
  - `passwordHash` (string)
  - `createdAt` (string)
  - `updatedAt` (string)

### 更新用户
- **接口**: `UpdateUser`
- **方法**: `PUT`
- **路径**: `/users/:id`
- **请求**: `UpdateUserReq`
  - `id` (int64, path): 用户 ID
  - `username` (string): 新用户名
- **响应**: `BaseResp`
  - `code` (int32)
  - `msg` (string)

### 获取用户列表
- **接口**: `ListUsers`
- **方法**: `GET`
- **路径**: `/users`
- **请求**: `PageReq`
  - `page` (int64, default=1): 页码
  - `pageSize` (int64, default=20): 每页数量
- **响应**: `UserListResp`
  - `list` ([]UserInfoResp): 用户列表
  - `page` (PageResp): 分页信息

---

## 项目 (Projects)

### 创建项目
- **接口**: `CreateProject`
- **方法**: `POST`
- **路径**: `/projects`
- **请求**: `CreateProjectReq`
  - `userId` (int64): 创建者 ID
  - `name` (string): 项目名称
  - `description` (string): 项目描述
- **响应**: `ProjectResp`
  - `id` (int64)
  - `name` (string)
  - `description` (string)
  - `ownerId` (int64)
  - `status` (string): active | archived
  - `createdAt` (string)
  - `updatedAt` (string)

### 删除项目
- **接口**: `DeleteProject`
- **方法**: `DELETE`
- **路径**: `/projects/:id`
- **请求**: `DeleteProjectReq`
  - `id` (int64, path): 项目 ID
- **响应**: `BaseResp`

### 更新项目
- **接口**: `UpdateProject`
- **方法**: `PUT`
- **路径**: `/projects/:id`
- **请求**: `UpdateProjectReq`
  - `id` (int64, path): 项目 ID
  - `name` (string)
  - `description` (string)
  - `status` (string)
- **响应**: `BaseResp`

### 获取项目列表
- **接口**: `ListProjects`
- **方法**: `GET`
- **路径**: `/projects`
- **请求**: `PageReq`
- **响应**: `ProjectListResp`
  - `list` ([]ProjectResp)
  - `page` (PageResp)

### 获取项目详情
- **接口**: `GetProject`
- **方法**: `GET`
- **路径**: `/projects/:id`
- **请求**: `GetProjectReq`
  - `id` (int64, path): 项目 ID
- **响应**: `ProjectResp`

### 邀请成员
- **接口**: `InviteMember`
- **方法**: `POST`
- **路径**: `/projects/:id/invite`
- **请求**: `InviteMemberReq`
  - `id` (int64, path): 项目 ID
  - `userId` (int64): 发起者 ID
  - `invitedUserId` (int64): 被邀请者 ID
  - `role` (string): owner | admin | developer | viewer
- **响应**: `BaseResp`

---

## 文件 (Files)

### 预上传文件
- **接口**: `PreUploadFile`
- **方法**: `POST`
- **路径**: `/files/preupload`
- **请求**: `PreUploadReq`
  - `projectId` (int64)
  - `name` (string)
  - `fileCategory` (string): text | image | video | audio | binary
  - `sizeBytes` (int64)
  - `hash` (string)
  - `mimeType` (string)
- **响应**: `PreUploadResp`
  - `uploadUrl` (string)
  - `fileId` (int64)
  - `versionId` (int64)
  - `versionNumber` (int64)

### 获取项目文件列表
- **接口**: `ListProjectFiles`
- **方法**: `GET`
- **路径**: `/projects/:projectId/files`
- **请求**: `ListProjectFilesReq`
  - `projectId` (int64, path)
  - `page` (int64)
  - `pageSize` (int64)
- **响应**: `ProjectFileListResp`
  - `list` ([]ProjectFileItem)
  - `page` (PageResp)

### 获取文件版本列表
- **接口**: `ListFileVersions`
- **方法**: `GET`
- **路径**: `/files/:id/versions`
- **请求**: `ListFileVersionsReq`
  - `id` (int64, path): 文件 ID
  - `page` (int64)
  - `pageSize` (int64)
- **响应**: `FileVersionListResp`
  - `list` ([]FileVersionItem)
  - `page` (PageResp)
