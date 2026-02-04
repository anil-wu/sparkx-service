// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/anil-wu/spark-x/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// md5 hash
	sum := md5.Sum([]byte(req.Password))
	passHash := hex.EncodeToString(sum[:])

	// try find by email
	user, findErr := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
	if findErr != nil && findErr != model.ErrNotFound {
		return nil, findErr
	}
	if user == nil {
		// register
		username := req.Email
		if idx := strings.Index(req.Email, "@"); idx > 0 {
			username = req.Email[:idx]
		}
		newUser := &model.Users{
			Username:     username,
			Email:        req.Email,
			PasswordHash: passHash,
		}
		_, err = l.svcCtx.UsersModel.Insert(l.ctx, newUser)
		if err != nil {
			return nil, err
		}
		return &types.LoginResp{
			UserId:  int64(newUser.Id),
			Created: true,
		}, nil
	}

	// check password
	if user.PasswordHash != passHash {
		return nil, model.InputParamInvalid
	}

	resp = &types.LoginResp{
		UserId:  int64(user.Id),
		Created: false,
	}

	return resp, nil
}
