// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSoftwareTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSoftwareTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSoftwareTemplateLogic {
	return &UpdateSoftwareTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSoftwareTemplateLogic) UpdateSoftwareTemplate(req *types.UpdateSoftwareTemplateReq) (resp *types.BaseResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("invalid template id")
	}

	// Check if template exists
	_, err = l.svcCtx.SoftwareTemplatesModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	// Build update data
	updateData := &model.SoftwareTemplates{}
	if req.Name != "" {
		updateData.Name = req.Name
	}
	if req.Description != "" {
		updateData.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.ArchiveFileId > 0 {
		updateData.ArchiveFileId = uint64(req.ArchiveFileId)
	}

	_, err = l.svcCtx.SoftwareTemplatesModel.Update(l.ctx, req.Id, updateData)
	if err != nil {
		return nil, err
	}

	resp = &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}

	return resp, nil
}
