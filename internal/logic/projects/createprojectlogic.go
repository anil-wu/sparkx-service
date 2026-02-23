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

	// 自动创建默认画布
	canvas := &model.WorkspaceCanvas{
		ProjectId:       p.Id,
		Name:            "Main Canvas",
		BackgroundColor: "#ffffff",
		CreatedBy:       uint64(req.UserId),
	}
	_, err = l.svcCtx.WorkspaceCanvasModel.Insert(l.ctx, canvas)
	if err != nil {
		l.Logger.Errorf("Failed to create default canvas for project %d: %v", p.Id, err)
		// 画布创建失败不影响项目创建，只记录日志
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
