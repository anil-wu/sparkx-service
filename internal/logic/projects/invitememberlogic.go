// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/anil-wu/spark-x/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInviteMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteMemberLogic {
	return &InviteMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteMemberLogic) InviteMember(req *types.InviteMemberReq) (resp *types.BaseResp, err error) {
	if req == nil || req.ProjectId <= 0 || req.InvitedUserId <= 0 || req.Role == "" {
		return nil, errors.New("invalid params")
	}
	pm := &model.ProjectMembers{
		ProjectId: uint64(req.ProjectId),
		UserId:    uint64(req.InvitedUserId),
		Role:      req.Role,
	}
	_, err = l.svcCtx.ProjectMembersModel.Insert(l.ctx, pm)
	if err != nil {
		return nil, err
	}
	resp = &types.BaseResp{Code: 0, Msg: "ok"}

	return resp, nil
}
