// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"errors"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadTemplateLogic {
	return &DownloadTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadTemplateLogic) DownloadTemplate(req *types.DownloadFileReq) (resp *types.DownloadFileResp, err error) {
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

	// 获取版本信息
	versionId := file.CurrentVersionId
	if req.VersionId > 0 {
		versionId = uint64(req.VersionId)
	}
	if req.VersionNumber > 0 {
		var v model.FileVersions
		if err := l.svcCtx.DB.WithContext(l.ctx).Where("file_id = ? AND version_number = ?", file.Id, req.VersionNumber).First(&v).Error; err != nil {
			return nil, err
		}
		versionId = v.Id
	}

	version, err := l.svcCtx.FileVersionsModel.FindOne(l.ctx, versionId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("file version not found")
		}
		return nil, err
	}
	if version.FileId != file.Id {
		return nil, errors.New("file version not found")
	}

	// 生成 OSS 临时访问 URL（GET 方法）
	url, err := l.svcCtx.OSSBucket.SignURL(version.StorageKey, "GET", int64(l.svcCtx.Config.OSS.ExpireSeconds))
	if err != nil {
		l.Errorf("[DownloadTemplate] Failed to sign URL: %v", err)
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(l.svcCtx.Config.OSS.ExpireSeconds) * time.Second).Format(time.RFC3339)
	l.Infof("[DownloadTemplate] Generated download URL for fileId=%d, versionId=%d, expiresAt=%s", req.Id, version.Id, expiresAt)

	resp = &types.DownloadFileResp{
		DownloadUrl: url,
		ExpiresAt:   expiresAt,
	}

	return resp, nil
}
