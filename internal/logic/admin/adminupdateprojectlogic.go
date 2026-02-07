package admin

import (
	"context"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateProjectLogic {
	return &AdminUpdateProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateProjectLogic) AdminUpdateProject(req *types.AdminUpdateProjectReq) (resp *types.BaseResp, err error) {
	project, err := l.svcCtx.ProjectsModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	// update name
	if req.Name != "" {
		project.Name = req.Name
	}

	// update description
	if req.Description != "" {
		project.Description.String = req.Description
	}

	// update status
	if req.Status != "" {
		if req.Status != "active" && req.Status != "archived" {
			return nil, model.InputParamInvalid
		}
		project.Status = req.Status
	}

	_, err = l.svcCtx.ProjectsModel.Update(l.ctx, req.Id, project)
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
