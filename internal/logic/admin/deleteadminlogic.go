package admin

import (
	"context"
	"encoding/json"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAdminLogic {
	return &DeleteAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAdminLogic) DeleteAdmin(req *types.DeleteAdminReq) (resp *types.BaseResp, err error) {
	// check if current admin is super_admin
	role := l.ctx.Value("role").(string)
	if role != "super_admin" {
		return nil, model.InputParamInvalid
	}

	// prevent delete self
	adminIdNumber, ok := l.ctx.Value("adminId").(json.Number)
	if ok {
		adminId, _ := adminIdNumber.Int64()
		if adminId == req.Id {
			return nil, model.InputParamInvalid
		}
	}

	u, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}
	if !u.IsSuper {
		return nil, model.InputParamInvalid
	}

	u.IsSuper = false
	_, err = l.svcCtx.UsersModel.Update(l.ctx, req.Id, u)
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
