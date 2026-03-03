package admin

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProfileLogic {
	return &AdminProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminProfileLogic) AdminProfile() (resp *types.AdminInfoResp, err error) {
	adminIdNumber, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	adminId, _ := adminIdNumber.Int64()

	u, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(adminId))
	if err != nil {
		return nil, err
	}

	role := "admin"
	if u.IsSuper {
		role = "super_admin"
	}

	return &types.AdminInfoResp{
		Id:          int64(u.Id),
		Username:    u.Email,
		Role:        role,
		Status:      "active",
		LastLoginAt: "",
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
	}, nil
}
