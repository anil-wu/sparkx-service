// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/anil-wu/spark-x/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProjectFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProjectFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProjectFilesLogic {
	return &ListProjectFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProjectFilesLogic) ListProjectFiles(req *types.ListProjectFilesReq) (resp *types.ProjectFileListResp, err error) {
	if req == nil || req.ProjectId <= 0 {
		return nil, errors.New("projectId required")
	}
	page := int(req.Page)
	size := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	var files []model.Files
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Files{}).Where("project_id = ?", req.ProjectId).Offset(offset).Limit(size).Order("id desc").Find(&files).Error; err != nil {
		return nil, err
	}
	items := make([]types.ProjectFileItem, 0, len(files))
	for _, f := range files {
		var latest model.FileVersions
		err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).Where("file_id = ?", f.Id).Order("version_number desc").Limit(1).Find(&latest).Error
		if err != nil {
			return nil, err
		}
		items = append(items, types.ProjectFileItem{
			Id:            int64(f.Id),
			ProjectId:     int64(f.ProjectId),
			Name:          f.Name,
			FileCategory:  f.FileCategory,
			VersionId:     int64(latest.Id),
			VersionNumber: int64(latest.VersionNumber),
			SizeBytes:     int64(latest.SizeBytes),
			Hash:          latest.Hash,
			MimeType:      latest.MimeType,
			CreatedAt:     f.CreatedAt.Format("2006-01-02 15:04:05"),
			StoragePath:   latest.StoragePath,
		})
	}
	var total int64
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Files{}).Where("project_id = ?", req.ProjectId).Count(&total).Error; err != nil {
		return nil, err
	}
	resp = &types.ProjectFileListResp{
		List: items,
		Page: types.PageResp{
			Page:     int64(page),
			PageSize: int64(size),
			Total:    total,
		},
	}

	return resp, nil
}
