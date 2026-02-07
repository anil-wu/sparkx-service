package admin

import (
	"context"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminDeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminDeleteUserLogic {
	return &AdminDeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminDeleteUserLogic) AdminDeleteUser(req *types.AdminDeleteUserReq) (resp *types.BaseResp, err error) {
	_, err = l.svcCtx.UsersModel.Delete(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
