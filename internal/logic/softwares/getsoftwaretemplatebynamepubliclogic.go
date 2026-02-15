// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package softwares

import (
	"context"
	"errors"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSoftwareTemplateByNamePublicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSoftwareTemplateByNamePublicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSoftwareTemplateByNamePublicLogic {
	return &GetSoftwareTemplateByNamePublicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSoftwareTemplateByNamePublicLogic) GetSoftwareTemplateByNamePublic(req *types.GetSoftwareTemplateByNameReq) (resp *types.SoftwareTemplateResp, err error) {
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not initialized")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("invalid template name")
	}

	var template model.SoftwareTemplates
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`name` = ?", name).First(&template).Error; err != nil {
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
