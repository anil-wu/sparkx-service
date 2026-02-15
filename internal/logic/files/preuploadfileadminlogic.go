// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreUploadFileAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreUploadFileAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreUploadFileAdminLogic {
	return &PreUploadFileAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreUploadFileAdminLogic) PreUploadFileAdmin(req *types.PreUploadReq) (resp *types.PreUploadResp, err error) {
	return NewPreUploadFileLogic(l.ctx, l.svcCtx).PreUploadFile(req)
}
