// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSoftwareTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSoftwareTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSoftwareTemplateLogic {
	return &DeleteSoftwareTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSoftwareTemplateLogic) DeleteSoftwareTemplate(req *types.DeleteSoftwareTemplateReq) (resp *types.BaseResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("invalid template id")
	}

	_, err = l.svcCtx.SoftwareTemplatesModel.Delete(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}

	resp = &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}

	return resp, nil
}
