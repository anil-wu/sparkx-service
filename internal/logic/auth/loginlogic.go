package auth

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
	"google.golang.org/api/idtoken"
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
	switch req.LoginType {
	case "google":
		return l.loginByGoogle(req)
	default:
		// default to email login for backward compatibility or explicit "email" type
		return l.loginByEmail(req)
	}
}

func (l *LoginLogic) loginByEmail(req *types.LoginReq) (*types.LoginResp, error) {
	if req.Email == "" || req.Password == "" {
		return nil, model.InputParamInvalid
	}

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
		_, err := l.svcCtx.UsersModel.Insert(l.ctx, newUser)
		if err != nil {
			return nil, err
		}

		// generate token
		token, err := l.generateToken(int64(newUser.Id))
		if err != nil {
			return nil, err
		}

		return &types.LoginResp{
			UserId:  int64(newUser.Id),
			Created: true,
			Token:   token,
		}, nil
	}

	// check password
	if user.PasswordHash != passHash {
		return nil, model.InputParamInvalid
	}

	// generate token
	token, err := l.generateToken(int64(user.Id))
	if err != nil {
		return nil, err
	}

	return &types.LoginResp{
		UserId:  int64(user.Id),
		Created: false,
		Token:   token,
	}, nil
}

func (l *LoginLogic) loginByGoogle(req *types.LoginReq) (*types.LoginResp, error) {
	if req.IdToken == "" {
		return nil, model.InputParamInvalid
	}

	// 1. Verify Google Token
	payload, err := idtoken.Validate(l.ctx, req.IdToken, l.svcCtx.Config.Google.ClientID)
	if err != nil {
		l.Logger.Errorf("Google token validation failed: %v", err)
		return nil, model.InputParamInvalid
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		l.Logger.Errorf("Google token missing email claim")
		return nil, model.InputParamInvalid
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	l.Logger.Infof(">>> Google User Info: Name=[%s], Email=[%s], Picture=[%s]", name, email, picture)

	// optional: check email_verified
	if verified, ok := payload.Claims["email_verified"].(bool); ok && !verified {
		l.Logger.Errorf("Google email not verified")
		return nil, model.InputParamInvalid
	}

	// 2. Find Identity or Create User
	// Check if identity exists
	provider := "google"
	providerUid := payload.Subject // "sub" claim

	identity, err := l.svcCtx.UserIdentitiesModel.FindOneByProviderProviderUid(l.ctx, provider, providerUid)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	var user *model.Users
	created := false

	if identity != nil {
		// Identity exists, find user
		user, err = l.svcCtx.UsersModel.FindOne(l.ctx, identity.UserId)
		if err != nil {
			return nil, err
		}
		// Update avatar if changed
		if picture != "" && user.Avatar != picture {
			user.Avatar = picture
			l.svcCtx.UsersModel.Update(l.ctx, int64(user.Id), user)
		}
	} else {
		// Identity not found, check if email exists in users table
		user, err = l.svcCtx.UsersModel.FindOneByEmail(l.ctx, email)
		if err != nil && err != model.ErrNotFound {
			return nil, err
		}

		if user == nil {
			// Create new user
			username := email
			if idx := strings.Index(email, "@"); idx > 0 {
				username = email[:idx]
			}
			if name != "" {
				username = name
			}

			newUser := &model.Users{
				Username:     username,
				Email:        email,
				PasswordHash: "", // No password for social login
				Avatar:       picture,
			}
			_, err := l.svcCtx.UsersModel.Insert(l.ctx, newUser)
			if err != nil {
				return nil, err
			}
			user = newUser
			created = true
		} else {
			// User exists (via email), update avatar if needed
			if picture != "" && user.Avatar != picture {
				user.Avatar = picture
				l.svcCtx.UsersModel.Update(l.ctx, int64(user.Id), user)
			}
		}

		// Create identity
		newIdentity := &model.UserIdentities{
			UserId:      user.Id,
			Provider:    provider,
			ProviderUid: providerUid,
			Email:       email,
		}
		_, err = l.svcCtx.UserIdentitiesModel.Insert(l.ctx, newIdentity)
		if err != nil {
			return nil, err
		}
	}

	// 3. Generate JWT
	token, err := l.generateToken(int64(user.Id))
	if err != nil {
		return nil, err
	}

	return &types.LoginResp{
		UserId:  int64(user.Id),
		Created: created,
		Token:   token,
	}, nil
}

func (l *LoginLogic) generateToken(userId int64) (string, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	secretKey := l.svcCtx.Config.Auth.AccessSecret

	claims := make(jwt.MapClaims)
	claims["exp"] = now + accessExpire
	claims["iat"] = now
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
