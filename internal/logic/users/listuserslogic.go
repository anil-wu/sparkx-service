// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package users

import (
	"context"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUsersLogic) ListUsers(req *types.PageReq) (resp *types.UserListResp, err error) {
	page := int(req.Page)
	size := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	var list []model.Users
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Users{}).Offset(offset).Limit(size).Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	var total int64
	if err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.Users{}).Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]types.UserInfoResp, 0, len(list))
	for _, u := range list {
		items = append(items, types.UserInfoResp{
			Id:           int64(u.Id),
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: u.PasswordHash,
			CreatedAt:    u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	resp = &types.UserListResp{
		List: items,
		Page: types.PageResp{
			Page:     int64(page),
			PageSize: int64(size),
			Total:    total,
		},
	}

	return resp, nil
}
