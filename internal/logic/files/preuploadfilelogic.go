// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// generateStoragePath 生成 OSS 存储路径
// 格式: 用户ID/项目ID/assets/hash前四位_hash后四位.后缀
func generateStoragePath(projectId, userId int64, fileCategory, fileHash, fileName string) string {
	// 取 hash 的前四位和后四位
	var shortHash string
	if len(fileHash) >= 8 {
		shortHash = fileHash[:4] + "_" + fileHash[len(fileHash)-4:]
	} else {
		shortHash = fileHash
	}

	// 清理文件名，获取后缀
	cleanName := filepath.Base(fileName)
	ext := filepath.Ext(cleanName)

	// 构建路径: userId/projectId/assets/hash.ext
	newFileName := shortHash + ext
	return fmt.Sprintf("%d/%d/assets/%s", userId, projectId, newFileName)
}

// getContentTypeByFormat 根据文件格式返回对应的 Content-Type
func getContentTypeByFormat(format string) string {
	switch strings.ToLower(format) {
	case "txt":
		return "text/plain"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "mp4":
		return "video/mp4"
	case "mp3":
		return "audio/mpeg"
	case "pdf":
		return "application/pdf"
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	case "zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

func normalizeOSSEndpointForSign(rawEndpoint, bucket string) (scheme string, host string, err error) {
	raw := strings.TrimSpace(rawEndpoint)
	if raw == "" {
		return "", "", errors.New("empty OSS endpoint")
	}

	u, parseErr := url.Parse(raw)
	if parseErr == nil && u.Host != "" {
		scheme = strings.TrimSpace(u.Scheme)
		host = strings.TrimSpace(u.Host)
	} else {
		u2, parseErr2 := url.Parse("https://" + raw)
		if parseErr2 != nil || u2.Host == "" {
			return "", "", errors.New("invalid OSS endpoint")
		}
		scheme = "https"
		host = strings.TrimSpace(u2.Host)
	}

	b := strings.TrimSpace(bucket)
	if b != "" {
		prefix := b + "."
		for strings.HasPrefix(host, prefix) {
			host = strings.TrimPrefix(host, prefix)
		}
	}

	if scheme == "" {
		scheme = "https"
	}
	if host == "" {
		return "", "", errors.New("invalid OSS endpoint host")
	}
	return scheme, host, nil
}

func escapeOSSObjectKeyPath(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func presignOSSPutURL(endpoint, bucket, accessKeyId, accessKeySecret, objectKey, contentType string, expireSeconds int64) (string, error) {
	if bucket == "" || accessKeyId == "" || accessKeySecret == "" || endpoint == "" {
		return "", errors.New("OSS not configured")
	}

	scheme, host, err := normalizeOSSEndpointForSign(endpoint, bucket)
	if err != nil {
		return "", err
	}

	expires := time.Now().Unix() + expireSeconds
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := "PUT\n\n" + contentType + "\n" + strconv.FormatInt(expires, 10) + "\n" + canonicalResource

	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	escapedObjectKey := escapeOSSObjectKeyPath(objectKey)
	return scheme + "://" + bucket + "." + host + "/" + escapedObjectKey +
		"?Expires=" + strconv.FormatInt(expires, 10) +
		"&OSSAccessKeyId=" + url.QueryEscape(accessKeyId) +
		"&Signature=" + url.QueryEscape(signature), nil
}

type PreUploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreUploadFileLogic {
	return &PreUploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreUploadFileLogic) PreUploadFile(req *types.PreUploadReq) (resp *types.PreUploadResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	isAdmin := false
	if !ok {
		adminIdNumber, ok2 := l.ctx.Value("adminId").(json.Number)
		if !ok2 {
			return nil, errors.New("unauthorized")
		}
		userIdNumber = adminIdNumber
		isAdmin = true
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Name == "" || req.FileCategory == "" || req.FileFormat == "" {
		return nil, model.InputParamInvalid
	}
	if req.ProjectId < 0 || (!isAdmin && req.ProjectId <= 0) {
		return nil, model.InputParamInvalid
	}

	if !isAdmin {
		var count int64
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).Where("project_id = ? AND user_id = ?", req.ProjectId, userId).Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("project not found or permission denied")
		}
	}

	// ensure file exists or create
	var file *model.Files
	// try find by name through project_files join
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Files{}).
		Joins("JOIN project_files ON project_files.file_id = files.id").
		Where("project_files.project_id = ? AND files.name = ?", req.ProjectId, req.Name).
		First(&file).Error
	if err != nil {
		// create new file
		newFile := &model.Files{
			Name:         req.Name,
			FileCategory: req.FileCategory,
			FileFormat:   req.FileFormat,
		}
		_, err = l.svcCtx.FilesModel.Insert(l.ctx, newFile)
		if err != nil {
			return nil, err
		}
		// create project_file relationship
		projectFile := &model.ProjectFiles{
			ProjectId: uint64(req.ProjectId),
			FileId:    newFile.Id,
		}
		_, err = l.svcCtx.ProjectFilesModel.Insert(l.ctx, projectFile)
		if err != nil {
			return nil, err
		}
		file = newFile
	}
	// get max version
	var maxVerNumber int64
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).Where("file_id = ?", file.Id).Select("COALESCE(MAX(version_number),0)").Scan(&maxVerNumber).Error
	if err != nil {
		return nil, err
	}
	nextVer := maxVerNumber + 1

	// 生成 OSS 存储路径: 项目ID/用户ID/文件类型/文件名_文件hash(简短).后缀
	objectPath := generateStoragePath(req.ProjectId, userId, req.FileCategory, req.Hash, req.Name)
	l.Infof("[PreUpload] ProjectId=%d, UserId=%d, File=%s, GeneratedPath=%s", req.ProjectId, userId, req.Name, objectPath)

	// create version row (without actual upload)
	newVer := &model.FileVersions{
		FileId:        file.Id,
		VersionNumber: uint64(nextVer),
		SizeBytes:     uint64(req.SizeBytes),
		Hash:          req.Hash,
		StorageKey:    objectPath,
		CreatedBy:     uint64(userId),
	}
	_, err = l.svcCtx.FileVersionsModel.Insert(l.ctx, newVer)
	if err != nil {
		l.Errorf("[PreUpload] Failed to insert version: %v", err)
		return nil, err
	}
	l.Infof("[PreUpload] Version inserted: VersionId=%d, StorageKey=%s", newVer.Id, objectPath)

	// update file current_version_id
	file.CurrentVersionId = newVer.Id
	_, err = l.svcCtx.FilesModel.Update(l.ctx, int64(file.Id), file)
	if err != nil {
		l.Errorf("[PreUpload] Failed to update file: %v", err)
		return nil, err
	}
	// 检查 OSS 是否已配置
	if l.svcCtx.OSSBucket == nil {
		l.Errorf("[PreUpload] OSS not configured")
		return nil, errors.New("OSS not configured")
	}

	// 确定 Content-Type
	contentType := req.ContentType
	if contentType == "" {
		// 根据文件格式推断 Content-Type
		contentType = getContentTypeByFormat(req.FileFormat)
	}

	url, err := presignOSSPutURL(
		l.svcCtx.Config.OSS.Endpoint,
		l.svcCtx.Config.OSS.Bucket,
		l.svcCtx.Config.OSS.AccessKeyId,
		l.svcCtx.Config.OSS.AccessKeySecret,
		objectPath,
		contentType,
		int64(l.svcCtx.Config.OSS.ExpireSeconds),
	)
	if err != nil {
		l.Errorf("[PreUpload] Failed to sign URL: %v", err)
		return nil, err
	}
	l.Infof("[PreUpload] Signed URL generated successfully, expires in %d seconds", l.svcCtx.Config.OSS.ExpireSeconds)
	resp = &types.PreUploadResp{
		UploadUrl:     url,
		FileId:        int64(file.Id),
		VersionId:     int64(newVer.Id),
		VersionNumber: int64(newVer.VersionNumber),
		ContentType:   contentType,
	}

	return resp, nil
}
