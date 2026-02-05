// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProjectLogic {
	return &CreateProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProjectLogic) CreateProject(req *types.CreateProjectReq) (resp *types.ProjectResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if ok {
		uid, _ := userIdNumber.Int64()
		req.UserId = uid
	}

	if req == nil || req.UserId <= 0 || req.Name == "" {
		return nil, errors.New("invalid params")
	}
	p := &model.Projects{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		OwnerId:     uint64(req.UserId),
		Status:      "active",
	}
	_, err = l.svcCtx.ProjectsModel.Insert(l.ctx, p)
	if err != nil {
		return nil, err
	}
	pm := &model.ProjectMembers{
		ProjectId: p.Id,
		UserId:    uint64(req.UserId),
		Role:      "owner",
	}
	_, err = l.svcCtx.ProjectMembersModel.Insert(l.ctx, pm)
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
