// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

import (
	"context"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSoftwareTemplatesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSoftwareTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSoftwareTemplatesLogic {
	return &ListSoftwareTemplatesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSoftwareTemplatesLogic) ListSoftwareTemplates(req *types.PageReq) (resp *types.SoftwareTemplateListResp, err error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var templates []model.SoftwareTemplates
	var total int64

	offset := (page - 1) * pageSize

	// Get total count
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SoftwareTemplates{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated list
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&templates).Error; err != nil {
		return nil, err
	}

	list := make([]types.SoftwareTemplateResp, 0, len(templates))
	for _, t := range templates {
		list = append(list, types.SoftwareTemplateResp{
			Id:            int64(t.Id),
			Name:          t.Name,
			Description:   t.Description.String,
			ArchiveFileId: int64(t.ArchiveFileId),
			CreatedBy:     int64(t.CreatedBy),
			CreatedAt:     t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	resp = &types.SoftwareTemplateListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}

	return resp, nil
}
