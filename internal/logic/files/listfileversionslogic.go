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

type ListFileVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFileVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFileVersionsLogic {
	return &ListFileVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFileVersionsLogic) ListFileVersions(req *types.ListFileVersionsReq) (resp *types.FileVersionListResp, err error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("file id required")
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
	var list []model.FileVersions
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).Where("file_id = ?", req.Id).Offset(offset).Limit(size).Order("version_number desc").Find(&list).Error; err != nil {
		return nil, err
	}
	var total int64
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).Where("file_id = ?", req.Id).Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]types.FileVersionItem, 0, len(list))
	for _, v := range list {
		items = append(items, types.FileVersionItem{
			Id:            int64(v.Id),
			FileId:        int64(v.FileId),
			VersionNumber: int64(v.VersionNumber),
			SizeBytes:     int64(v.SizeBytes),
			Hash:          v.Hash,
			StoragePath:   v.StoragePath,
			MimeType:      v.MimeType,
			CreatedAt:     v.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     v.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreatedBy:     int64(v.CreatedBy),
		})
	}
	resp = &types.FileVersionListResp{
		List: items,
		Page: types.PageResp{
			Page:     int64(page),
			PageSize: int64(size),
			Total:    total,
		},
	}

	return resp, nil
}
