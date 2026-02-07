// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFileContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileContentLogic {
	return &GetFileContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetFileContent 获取文件内容（代理访问）
// 直接返回文件内容，适合图片预览等场景
func (l *GetFileContentLogic) GetFileContent(req *types.GetFileContentReq) (io.ReadCloser, string, error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, "", errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	// 获取文件信息
	var file model.Files
	err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Files{}).
		Where("id = ? AND deleted_at IS NULL", req.Id).
		First(&file).Error
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, "", errors.New("file not found")
		}
		return nil, "", err
	}

	// Get project_id from project_files
	var projectFile model.ProjectFiles
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("file_id = ?", file.Id).First(&projectFile).Error; err != nil {
		return nil, "", err
	}

	// 检查用户是否有权限访问该项目的文件
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
		Where("project_id = ? AND user_id = ?", projectFile.ProjectId, userId).
		Count(&count).Error; err != nil {
		return nil, "", err
	}
	if count == 0 {
		return nil, "", errors.New("project not found or permission denied")
	}

	// 获取当前版本信息
	var version model.FileVersions
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).
		Where("id = ?", file.CurrentVersionId).
		First(&version).Error
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, "", errors.New("file version not found")
		}
		return nil, "", err
	}

	// 从 OSS 获取文件内容
	reader, err := l.svcCtx.OSSBucket.GetObject(version.StorageKey)
	if err != nil {
		l.Errorf("[GetFileContent] Failed to get object from OSS: %v", err)
		return nil, "", err
	}

	// 获取 Content-Type
	contentType := getContentTypeByFormat(file.FileFormat)

	l.Infof("[GetFileContent] Serving file content for fileId=%d, versionId=%d, contentType=%s",
		req.Id, version.Id, contentType)

	return reader, contentType, nil
}
