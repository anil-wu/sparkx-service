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

type UpdateProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProjectLogic {
	return &UpdateProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProjectLogic) UpdateProject(req *types.UpdateProjectReq) (resp *types.BaseResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, errors.New("id required")
	}

	// Check ownership
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Projects{}).Where("id = ? AND owner_id = ?", req.Id, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	data := &model.Projects{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      req.Status,
	}
	_, err = l.svcCtx.ProjectsModel.Update(l.ctx, req.Id, data)
	if err != nil {
		return nil, err
	}
	resp = &types.BaseResp{Code: 0, Msg: "ok"}

	return resp, nil
}
