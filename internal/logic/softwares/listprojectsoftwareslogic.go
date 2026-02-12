package softwares

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProjectSoftwaresLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProjectSoftwaresLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProjectSoftwaresLogic {
	return &ListProjectSoftwaresLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProjectSoftwaresLogic) ListProjectSoftwares(req *types.ListProjectSoftwaresReq) (resp *types.SoftwareListResp, err error) {
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

	if req == nil || req.ProjectId <= 0 {
		return nil, model.InputParamInvalid
	}
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
	}

	if !isAdmin {
		var count int64
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
			Where("project_id = ? AND user_id = ?", req.ProjectId, userId).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("project not found or permission denied")
		}
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

	var softwares []model.Softwares
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Softwares{}).
		Where("project_id = ?", req.ProjectId).
		Offset(offset).
		Limit(size).
		Order("id desc").
		Find(&softwares).Error; err != nil {
		return nil, err
	}

	items := make([]types.SoftwareItem, 0, len(softwares))
	for _, sw := range softwares {
		items = append(items, types.SoftwareItem{
			Id:              int64(sw.Id),
			ProjectId:       int64(sw.ProjectId),
			Name:            sw.Name,
			Description:     sw.Description.String,
			TemplateId:      int64(sw.TemplateId),
			TechnologyStack: sw.TechnologyStack,
			Status:          sw.Status,
			CreatedBy:       int64(sw.CreatedBy),
			CreatedAt:       sw.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       sw.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	var total int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Softwares{}).
		Where("project_id = ?", req.ProjectId).
		Count(&total).Error; err != nil {
		return nil, err
	}

	return &types.SoftwareListResp{
		List: items,
		Page: types.PageResp{
			Page:     int64(page),
			PageSize: int64(size),
			Total:    total,
		},
	}, nil
}
