package admin

import (
	"context"
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
	adminId := l.ctx.Value("adminId").(int64)

	admin, err := l.svcCtx.AdminsModel.FindOne(l.ctx, uint64(adminId))
	if err != nil {
		return nil, err
	}

	lastLoginAt := ""
	if !admin.LastLoginAt.IsZero() {
		lastLoginAt = admin.LastLoginAt.Format(time.RFC3339)
	}

	return &types.AdminInfoResp{
		Id:          int64(admin.Id),
		Username:    admin.Username,
		Role:        admin.Role,
		Status:      admin.Status,
		LastLoginAt: lastLoginAt,
		CreatedAt:   admin.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   admin.UpdatedAt.Format(time.RFC3339),
	}, nil
}
