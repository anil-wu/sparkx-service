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

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Users{}).Where("`is_super` = TRUE")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.Users
	if err := query.Order("`id` DESC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&users).Error; err != nil {
		return nil, err
	}

	list := make([]types.AdminInfoResp, 0, len(users))
	for _, u := range users {
		list = append(list, types.AdminInfoResp{
			Id:          int64(u.Id),
			Username:    u.Email,
			Role:        "super_admin",
			Status:      "active",
			LastLoginAt: "",
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.AdminListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}
