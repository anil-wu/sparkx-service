// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workspace

import (
	"context"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreLayerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreLayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreLayerLogic {
	return &RestoreLayerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreLayerLogic) RestoreLayer(req *types.RestoreLayerReq) (resp *types.RestoreLayerResp, err error) {
	// 恢复已删除的图层
	affected, err := l.svcCtx.WorkspaceLayerModel.Restore(l.ctx, uint64(req.Id))
	if err != nil {
		l.Logger.Errorf("restore layer error: %v", err)
		return nil, err
	}

	if affected == 0 {
		l.Logger.Errorf("layer not found or not deleted: %d", req.Id)
		return nil, model.ErrNotFound
	}

	restoredAt := time.Now().Format("2006-01-02 15:04:05")

	return &types.RestoreLayerResp{
		LayerId:    req.Id,
		Restored:   true,
		RestoredAt: restoredAt,
	}, nil
}
