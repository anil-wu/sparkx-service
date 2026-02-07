package admin

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAdminLogic {
	return &UpdateAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAdminLogic) UpdateAdmin(req *types.UpdateAdminReq) (resp *types.BaseResp, err error) {
	// check if current admin is super_admin
	role := l.ctx.Value("role").(string)
	if role != "super_admin" {
		return nil, model.InputParamInvalid
	}

	admin, err := l.svcCtx.AdminsModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	// update password
	if req.Password != "" {
		sum := md5.Sum([]byte(req.Password))
		admin.PasswordHash = hex.EncodeToString(sum[:])
	}

	// update role
	if req.Role != "" {
		if req.Role != "super_admin" && req.Role != "admin" {
			return nil, model.InputParamInvalid
		}
		admin.Role = req.Role
	}

	// update status
	if req.Status != "" {
		if req.Status != "active" && req.Status != "disabled" {
			return nil, model.InputParamInvalid
		}
		admin.Status = req.Status
	}

	_, err = l.svcCtx.AdminsModel.Update(l.ctx, req.Id, admin)
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}, nil
}
