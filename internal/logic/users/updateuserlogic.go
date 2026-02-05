// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package users

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (resp *types.BaseResp, err error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("id required")
	}

	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if ok {
		uid, _ := userIdNumber.Int64()
		if uid != req.Id {
			return nil, errors.New("permission denied")
		}
	}

	data := &model.Users{
		Username: req.Username,
	}
	_, err = l.svcCtx.UsersModel.Update(l.ctx, req.Id, data)
	if err != nil {
		return nil, err
	}
	resp = &types.BaseResp{Code: 0, Msg: "ok"}

	return resp, nil
}
