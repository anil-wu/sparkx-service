package admin_auth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (resp *types.AdminLoginResp, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, model.InputParamInvalid
	}

	// md5 hash
	sum := md5.Sum([]byte(req.Password))
	passHash := hex.EncodeToString(sum[:])

	// find admin by username
	admin, findErr := l.svcCtx.AdminsModel.FindOneByUsername(l.ctx, req.Username)
	if findErr != nil {
		if findErr == model.ErrNotFound {
			return nil, model.InputParamInvalid
		}
		return nil, findErr
	}

	// check password
	if admin.PasswordHash != passHash {
		return nil, model.InputParamInvalid
	}

	// check status
	if admin.Status != "active" {
		return nil, model.InputParamInvalid
	}

	// update last login time
	admin.LastLoginAt = time.Now()
	_, _ = l.svcCtx.AdminsModel.Update(l.ctx, int64(admin.Id), admin)

	// generate token
	token, err := l.generateToken(int64(admin.Id), admin.Role)
	if err != nil {
		return nil, err
	}

	return &types.AdminLoginResp{
		AdminId: int64(admin.Id),
		Role:    admin.Role,
		Token:   token,
	}, nil
}

func (l *AdminLoginLogic) generateToken(adminId int64, role string) (string, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.AdminAuth.AccessExpire
	accessSecret := l.svcCtx.Config.AdminAuth.AccessSecret

	claims := jwt.MapClaims{
		"adminId": adminId,
		"role":    role,
		"iat":     now,
		"exp":     now + accessExpire,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}
