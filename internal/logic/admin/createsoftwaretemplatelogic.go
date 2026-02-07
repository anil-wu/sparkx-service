// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

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

type CreateSoftwareTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSoftwareTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSoftwareTemplateLogic {
	return &CreateSoftwareTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSoftwareTemplateLogic) CreateSoftwareTemplate(req *types.CreateSoftwareTemplateReq) (resp *types.SoftwareTemplateResp, err error) {
	// Get admin ID from JWT context
	adminIdNumber, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	adminId, _ := adminIdNumber.Int64()

	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// Create software template
	template := &model.SoftwareTemplates{
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		ArchiveFileId: uint64(req.ArchiveFileId),
		CreatedBy:     uint64(adminId),
	}

	_, err = l.svcCtx.SoftwareTemplatesModel.Insert(l.ctx, template)
	if err != nil {
		return nil, err
	}

	resp = &types.SoftwareTemplateResp{
		Id:            int64(template.Id),
		Name:          template.Name,
		Description:   template.Description.String,
		ArchiveFileId: int64(template.ArchiveFileId),
		CreatedBy:     int64(template.CreatedBy),
		CreatedAt:     template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     template.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}
