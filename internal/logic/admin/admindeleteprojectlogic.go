package admin

import (
	"context"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeleteProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteProjectLogic {
	return &AdminDeleteProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeleteProjectLogic) AdminDeleteProject(req *types.AdminDeleteProjectReq) (resp *types.BaseResp, err error) {
	_, err = l.svcCtx.ProjectsModel.Delete(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
