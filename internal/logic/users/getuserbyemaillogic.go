// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package users

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByEmailLogic {
	return &GetUserByEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserByEmailLogic) GetUserByEmail(req *types.GetUserByEmailReq) (resp *types.UserInfoResp, err error) {
	if req == nil || req.Email == "" {
		return nil, errors.New("email required")
	}
	u, err := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		return nil, err
	}
	resp = &types.UserInfoResp{
		Id:           int64(u.Id),
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}
