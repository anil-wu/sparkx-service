// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadFileLogic {
	return &DownloadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadFileLogic) DownloadFile(req *types.DownloadFileReq) (resp *types.DownloadFileResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	// 获取文件信息
	file, err := l.svcCtx.FilesModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	// Get project_id from project_files
	var projectFile model.ProjectFiles
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("file_id = ?", file.Id).First(&projectFile).Error; err != nil {
		return nil, err
	}

	// 检查项目成员权限
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
		Where("project_id = ? AND user_id = ?", projectFile.ProjectId, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	// 获取当前版本信息
	version, err := l.svcCtx.FileVersionsModel.FindOne(l.ctx, file.CurrentVersionId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("file version not found")
		}
		return nil, err
	}

	// 生成 OSS 临时访问 URL（GET 方法）
	// 不指定 Content-Type，避免签名不匹配
	url, err := l.svcCtx.OSSBucket.SignURL(version.StorageKey, "GET", int64(l.svcCtx.Config.OSS.ExpireSeconds))
	if err != nil {
		l.Errorf("[DownloadFile] Failed to sign URL: %v", err)
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(l.svcCtx.Config.OSS.ExpireSeconds) * time.Second).Format(time.RFC3339)
	l.Infof("[DownloadFile] Generated download URL for fileId=%d, versionId=%d, expiresAt=%s", req.Id, version.Id, expiresAt)

	resp = &types.DownloadFileResp{
		DownloadUrl: url,
		ExpiresAt:   expiresAt,
	}

	return resp, nil
}
