package admin

import (
	"context"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUsersLogic {
	return &AdminListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListUsersLogic) AdminListUsers(req *types.PageReq) (resp *types.UserListResp, err error) {
	var users []model.Users
	result := l.svcCtx.DB.Order("id DESC").Offset(int((req.Page - 1) * req.PageSize)).Limit(int(req.PageSize)).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	// get total count
	var total int64
	l.svcCtx.DB.Model(&model.Users{}).Count(&total)

	list := make([]types.UserInfoResp, 0, len(users))
	for _, user := range users {
		list = append(list, types.UserInfoResp{
			Id:           int64(user.Id),
			Username:     user.Username,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			CreatedAt:    user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.UserListResp{
		List: list,
		Page: types.PageResp{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
