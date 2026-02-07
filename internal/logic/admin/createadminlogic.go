package admin

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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

	// check if username already exists
	_, err = l.svcCtx.AdminsModel.FindOneByUsername(l.ctx, req.Username)
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

	newAdmin := &model.Admins{
		Username:     req.Username,
		PasswordHash: passHash,
		Role:         adminRole,
		Status:       "active",
	}

	_, err = l.svcCtx.AdminsModel.Insert(l.ctx, newAdmin)
	if err != nil {
		return nil, err
	}

	return &types.AdminInfoResp{
		Id:        int64(newAdmin.Id),
		Username:  newAdmin.Username,
		Role:      newAdmin.Role,
		Status:    newAdmin.Status,
		CreatedAt: newAdmin.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newAdmin.UpdatedAt.Format(time.RFC3339),
	}, nil
}
