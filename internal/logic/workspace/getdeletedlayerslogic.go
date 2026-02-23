// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workspace

import (
	"context"

	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeletedLayersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeletedLayersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeletedLayersLogic {
	return &GetDeletedLayersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeletedLayersLogic) GetDeletedLayers(req *types.GetDeletedLayersReq) (resp *types.DeletedLayersResp, err error) {
	// 查找已删除的图层
	deletedLayers, err := l.svcCtx.WorkspaceLayerModel.FindDeletedByCanvasId(l.ctx, uint64(req.CanvasId), req.Limit)
	if err != nil {
		l.Logger.Errorf("find deleted layers error: %v", err)
		return nil, err
	}

	resp = &types.DeletedLayersResp{
		DeletedLayers: []types.DeletedLayerInfo{},
		Total:         int64(len(deletedLayers)),
	}

	for _, layer := range deletedLayers {
		layerInfo := types.DeletedLayerInfo{
			Id:        int64(layer.Id),
			Name:      layer.Name,
			LayerType: layer.LayerType,
		}

		if layer.DeletedAt.Valid {
			layerInfo.DeletedAt = layer.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}

		if layer.DeletedBy.Valid {
			layerInfo.DeletedBy = layer.DeletedBy.Int64
		}

		resp.DeletedLayers = append(resp.DeletedLayers, layerInfo)
	}

	return
}
