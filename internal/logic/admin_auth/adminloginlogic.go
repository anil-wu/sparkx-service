package admin_auth

import (
	"context"
	"crypto/md5"
	"database/sql"
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

	// find admin by username
	admin, findErr := l.svcCtx.AdminsModel.FindOneByUsername(l.ctx, req.Username)
	if findErr != nil {
		if findErr == model.ErrNotFound {
			return nil, model.InputParamInvalid
		}
		return nil, findErr
	}

	// check password
	if !passwordMatches(admin.PasswordHash, req.Password) {
		return nil, model.InputParamInvalid
	}

	// check status
	if !statusAllowsLogin(admin.Status) {
		return nil, model.InputParamInvalid
	}

	// update last login time
	admin.LastLoginAt = sql.NullTime{Time: time.Now(), Valid: true}
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
