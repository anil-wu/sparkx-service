// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSoftwareTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSoftwareTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSoftwareTemplateLogic {
	return &GetSoftwareTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSoftwareTemplateLogic) GetSoftwareTemplate(req *types.GetSoftwareTemplateReq) (resp *types.SoftwareTemplateResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("invalid template id")
	}

	template, err := l.svcCtx.SoftwareTemplatesModel.FindOne(l.ctx, uint64(req.Id))
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
