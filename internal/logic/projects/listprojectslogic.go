// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProjectsLogic {
	return &ListProjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProjectsLogic) ListProjects(req *types.PageReq) (resp *types.ProjectListResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

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

	var list []model.Projects
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Projects{}).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userId).
		Offset(offset).Limit(size).Order("projects.id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	var total int64
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Projects{}).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userId).
		Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]types.ProjectResp, 0, len(list))
	for _, p := range list {
		items = append(items, types.ProjectResp{
			Id:          int64(p.Id),
			Name:        p.Name,
			Description: p.Description.String,
			OwnerId:     int64(p.OwnerId),
			Status:      p.Status,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	resp = &types.ProjectListResp{
		List: items,
		Page: types.PageResp{
			Page:     int64(page),
			PageSize: int64(size),
			Total:    total,
		},
	}

	return resp, nil
}
