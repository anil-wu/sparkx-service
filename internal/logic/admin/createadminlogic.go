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

type CreateAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAdminLogic {
	return &CreateAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAdminLogic) CreateAdmin(req *types.CreateAdminReq) (resp *types.AdminInfoResp, err error) {
	// check if current admin is super_admin
	role := l.ctx.Value("role").(string)
	if role != "super_admin" {
		return nil, model.InputParamInvalid
	}

	if req.Username == "" || req.Password == "" {
		return nil, model.InputParamInvalid
	}

	email := strings.TrimSpace(req.Username)
	_, err = l.svcCtx.UsersModel.FindOneByEmail(l.ctx, email)
	if err == nil {
		return nil, model.InputParamInvalid
	}
	if err != model.ErrNotFound {
		return nil, err
	}

	// md5 hash
	sum := md5.Sum([]byte(req.Password))
	passHash := hex.EncodeToString(sum[:])

	// validate role
	adminRole := req.Role
	if adminRole == "" {
		adminRole = "admin"
	}
	if adminRole != "super_admin" && adminRole != "admin" {
		return nil, model.InputParamInvalid
	}

	username := email
	if idx := strings.Index(email, "@"); idx > 0 {
		username = email[:idx]
	}

	newUser := &model.Users{
		Username:     username,
		Email:        email,
		PasswordHash: passHash,
		IsSuper:      adminRole == "super_admin",
	}

	_, err = l.svcCtx.UsersModel.Insert(l.ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &types.AdminInfoResp{
		Id:        int64(newUser.Id),
		Username:  newUser.Email,
		Role:      adminRole,
		Status:    "active",
		CreatedAt: newUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}
