// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"context"
	"errors"

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
	if req == nil || req.Id <= 0 {
		return nil, errors.New("id required")
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
