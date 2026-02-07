package admin

import (
	"context"
	"database/sql"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateProjectLogic {
	return &AdminCreateProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateProjectLogic) AdminCreateProject(req *types.AdminCreateProjectReq) (resp *types.ProjectResp, err error) {
	if req.Name == "" || req.OwnerId == 0 {
		return nil, model.InputParamInvalid
	}

	// check if owner exists
	_, err = l.svcCtx.UsersModel.FindOne(l.ctx, uint64(req.OwnerId))
	if err != nil {
		return nil, err
	}

	newProject := &model.Projects{
		Name: req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		OwnerId: uint64(req.OwnerId),
		Status:  "active",
	}

	_, err = l.svcCtx.ProjectsModel.Insert(l.ctx, newProject)
	if err != nil {
		return nil, err
	}

	// add owner as project member
	member := &model.ProjectMembers{
		ProjectId: newProject.Id,
		UserId:    uint64(req.OwnerId),
		Role:      "owner",
	}
	_, _ = l.svcCtx.ProjectMembersModel.Insert(l.ctx, member)

	return &types.ProjectResp{
		Id:          int64(newProject.Id),
		Name:        newProject.Name,
		Description: newProject.Description.String,
		CoverFileId: int64(newProject.CoverFileId),
		OwnerId:     int64(newProject.OwnerId),
		Status:      newProject.Status,
		CreatedAt:   newProject.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   newProject.UpdatedAt.Format(time.RFC3339),
	}, nil
}
