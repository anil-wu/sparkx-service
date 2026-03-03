package admin_auth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
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

	user, findErr := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, strings.TrimSpace(req.Username))
	if findErr != nil {
		if findErr == model.ErrNotFound {
			return nil, model.InputParamInvalid
		}
		return nil, findErr
	}
	if user == nil || !user.IsSuper {
		return nil, model.InputParamInvalid
	}

	if !passwordMatches(user.PasswordHash, req.Password) {
		return nil, model.InputParamInvalid
	}

	token, err := l.generateToken(int64(user.Id), user.IsSuper)
	if err != nil {
		return nil, err
	}

	return &types.AdminLoginResp{
		AdminId: int64(user.Id),
		Role:    "super_admin",
		Token:   token,
	}, nil
}

func (l *AdminLoginLogic) generateToken(userId int64, isSuper bool) (string, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	accessSecret := l.svcCtx.Config.Auth.AccessSecret

	claims := jwt.MapClaims{
		"userId":  userId,
		"isSuper": isSuper,
		"iat":     now,
		"exp":     now + accessExpire,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}

func statusAllowsLogin(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "", "active", "enabled", "1", "true":
		return true
	case "disabled", "inactive", "0", "false":
		return false
	default:
		return false
	}
}

func passwordMatches(storedHashOrPassword string, inputPassword string) bool {
	stored := strings.TrimSpace(storedHashOrPassword)
	input := strings.TrimSpace(inputPassword)
	if stored == "" || input == "" {
		return false
	}

	if isHex32(stored) {
		if isHex32(input) {
			return strings.EqualFold(stored, input)
		}
		return strings.EqualFold(stored, md5HexLower(input))
	}

	if stored == input {
		return true
	}

	if isHex32(input) {
		return strings.EqualFold(stored, input)
	}

	return false
}

func md5HexLower(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func isHex32(value string) bool {
	if len(value) != 32 {
		return false
	}
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
