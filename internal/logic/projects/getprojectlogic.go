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

type GetProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProjectLogic {
	return &GetProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProjectLogic) GetProject(req *types.GetProjectReq) (resp *types.ProjectResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, errors.New("id required")
	}

	// Check project membership
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).Where("project_id = ? AND user_id = ?", req.Id, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	p, err := l.svcCtx.ProjectsModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}
	resp = &types.ProjectResp{
		Id:          int64(p.Id),
		Name:        p.Name,
		Description: p.Description.String,
		OwnerId:     int64(p.OwnerId),
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}
