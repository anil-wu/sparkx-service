package admin

import (
	"context"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdminsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAdminsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdminsLogic {
	return &ListAdminsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAdminsLogic) ListAdmins(req *types.ListAdminsReq) (resp *types.AdminListResp, err error) {
	// check if current admin is super_admin
	role := l.ctx.Value("role").(string)
	if role != "super_admin" {
		return nil, model.InputParamInvalid
	}

	// get all admins using GORM
	var admins []model.Admins
	result := l.svcCtx.DB.Order("id DESC").Offset(int((req.Page - 1) * req.PageSize)).Limit(int(req.PageSize)).Find(&admins)
	if result.Error != nil {
		return nil, result.Error
	}

	// get total count
	var total int64
	l.svcCtx.DB.Model(&model.Admins{}).Count(&total)

	list := make([]types.AdminInfoResp, 0, len(admins))
	for _, admin := range admins {
		lastLoginAt := ""
		if admin.LastLoginAt.Valid {
			lastLoginAt = admin.LastLoginAt.Time.Format(time.RFC3339)
		}
		list = append(list, types.AdminInfoResp{
			Id:          int64(admin.Id),
			Username:    admin.Username,
			Role:        admin.Role,
			Status:      admin.Status,
			LastLoginAt: lastLoginAt,
			CreatedAt:   admin.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   admin.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.AdminListResp{
		List: list,
		Page: types.PageResp{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
