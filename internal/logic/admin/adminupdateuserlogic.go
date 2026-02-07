package admin

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateUserLogic {
	return &AdminUpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateUserLogic) AdminUpdateUser(req *types.AdminUpdateUserReq) (resp *types.BaseResp, err error) {
	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	// update username
	if req.Username != "" {
		user.Username = req.Username
	}

	// update email
	if req.Email != "" {
		user.Email = req.Email
	}

	// update password
	if req.Password != "" {
		sum := md5.Sum([]byte(req.Password))
		user.PasswordHash = hex.EncodeToString(sum[:])
	}

	_, err = l.svcCtx.UsersModel.Update(l.ctx, req.Id, user)
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
