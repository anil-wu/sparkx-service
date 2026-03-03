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

type DeleteLayerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLayerLogic {
	return &DeleteLayerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLayerLogic) DeleteLayer(req *types.DeleteLayerReq) (resp *types.DeleteLayerResp, err error) {
	layer, err := l.svcCtx.WorkspaceLayerModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		return nil, err
	}
	canvas, err := l.svcCtx.WorkspaceCanvasModel.FindOne(l.ctx, layer.CanvasId)
	if err != nil {
		return nil, err
	}
	userId, err := ensureProjectMember(l.ctx, l.svcCtx, int64(canvas.ProjectId))
	if err != nil {
		return nil, err
	}

	// 执行软删除
	affected, err := l.svcCtx.WorkspaceLayerModel.SoftDelete(l.ctx, uint64(req.Id), uint64(userId))
	if err != nil {
		l.Logger.Errorf("soft delete layer error: %v", err)
		return nil, err
	}

	if affected == 0 {
		l.Logger.Errorf("layer not found or already deleted: %d", req.Id)
		return nil, model.ErrNotFound
	}

	deletedAt := time.Now().Format("2006-01-02 15:04:05")

	return &types.DeleteLayerResp{
		LayerId:   req.Id,
		Deleted:   true,
		DeletedAt: deletedAt,
	}, nil
}
