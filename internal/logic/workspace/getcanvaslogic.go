// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workspace

import (
	"context"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCanvasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCanvasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCanvasLogic {
	return &GetCanvasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCanvasLogic) GetCanvas(req *types.GetCanvasReq) (resp *types.CanvasWithLayersResp, err error) {
	// 查找画布
	canvas, err := l.svcCtx.WorkspaceCanvasModel.FindOneByProjectId(l.ctx, uint64(req.ProjectId))
	if err != nil && err != model.ErrNotFound {
		l.Logger.Errorf("find canvas by project id error: %v", err)
		return nil, err
	}

	resp = &types.CanvasWithLayersResp{
		Canvas: types.CanvasResp{},
		Layers: []types.LayerResp{},
	}

	// 如果画布存在，填充画布信息和图层
	if canvas != nil {
		resp.Canvas = types.CanvasResp{
			Id:              int64(canvas.Id),
			ProjectId:       int64(canvas.ProjectId),
			Name:            canvas.Name,
			BackgroundColor: canvas.BackgroundColor,
			CreatedAt:       canvas.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       canvas.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreatedBy:       int64(canvas.CreatedBy),
		}

		// 解析 Metadata
		if canvas.Metadata.Valid && canvas.Metadata.String != "" {
			// 这里可以解析 JSON，暂时简单处理
			resp.Canvas.Metadata = &types.CanvasMetadata{}
		}

		// 查找所有未删除的图层
		layers, err := l.svcCtx.WorkspaceLayerModel.FindByCanvasId(l.ctx, canvas.Id)
		if err != nil {
			l.Logger.Errorf("find layers by canvas id error: %v", err)
			return nil, err
		}

		// 转换图层数据（只包含未删除的图层）
		for _, layer := range layers {
			// 双重检查：确保不返回已删除的图层
			if layer.Deleted {
				continue
			}

			layerResp := types.LayerResp{
				Id:        int64(layer.Id),
				CanvasId:  int64(layer.CanvasId),
				LayerType: layer.LayerType,
				Name:      layer.Name,
				ZIndex:    int64(layer.ZIndex),
				PositionX: layer.PositionX,
				PositionY: layer.PositionY,
				Width:     layer.Width,
				Height:    layer.Height,
				Rotation:  layer.Rotation,
				Visible:   layer.Visible,
				Locked:    layer.Locked,
				Deleted:   layer.Deleted,
				CreatedAt: layer.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: layer.UpdatedAt.Format("2006-01-02 15:04:05"),
				CreatedBy: int64(layer.CreatedBy),
			}

			// 解析 Properties JSON
			if layer.Properties != "" {
				// 这里需要解析 JSON 到 LayerProperties，暂时留空
				layerResp.Properties = &types.LayerProperties{}
			}

			// 处理 FileId
			if layer.FileId.Valid {
				fileId := int64(layer.FileId.Int64)
				layerResp.FileId = &fileId
			}

			resp.Layers = append(resp.Layers, layerResp)
		}
	}

	return
}
