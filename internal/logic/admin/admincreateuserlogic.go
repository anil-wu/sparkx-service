package admin

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateUserLogic {
	return &AdminCreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateUserLogic) AdminCreateUser(req *types.AdminCreateUserReq) (resp *types.UserInfoResp, err error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, model.InputParamInvalid
	}

	// check if email already exists
	_, err = l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
	if err == nil {
		return nil, model.InputParamInvalid
	}
	if err != model.ErrNotFound {
		return nil, err
	}

	// md5 hash
	sum := md5.Sum([]byte(req.Password))
	passHash := hex.EncodeToString(sum[:])

	username := req.Username
	if username == "" {
		username = req.Email
		if idx := strings.Index(req.Email, "@"); idx > 0 {
			username = req.Email[:idx]
		}
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

	return &types.UserInfoResp{
		Id:           int64(newUser.Id),
		Username:     newUser.Username,
		Email:        newUser.Email,
		PasswordHash: newUser.PasswordHash,
		CreatedAt:    newUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    newUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}
