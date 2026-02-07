package admin

import (
	"context"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListProjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListProjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListProjectsLogic {
	return &AdminListProjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListProjectsLogic) AdminListProjects(req *types.PageReq) (resp *types.ProjectListResp, err error) {
	var projects []model.Projects
	result := l.svcCtx.DB.Order("id DESC").Offset(int((req.Page - 1) * req.PageSize)).Limit(int(req.PageSize)).Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}

	// get total count
	var total int64
	l.svcCtx.DB.Model(&model.Projects{}).Count(&total)

	list := make([]types.ProjectResp, 0, len(projects))
	for _, project := range projects {
		list = append(list, types.ProjectResp{
			Id:          int64(project.Id),
			Name:        project.Name,
			Description: project.Description.String,
			CoverFileId: int64(project.CoverFileId),
			OwnerId:     int64(project.OwnerId),
			Status:      project.Status,
			CreatedAt:   project.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.ProjectListResp{
		List: list,
		Page: types.PageResp{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
