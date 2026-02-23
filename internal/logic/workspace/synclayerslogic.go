// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workspace

import (
	"context"
	"encoding/json"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncLayersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncLayersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncLayersLogic {
	return &SyncLayersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncLayersLogic) SyncLayers(req *types.SyncLayersReq) (resp *types.SyncLayersResp, err error) {
	resp = &types.SyncLayersResp{
		Uploaded:     0,
		Updated:      0,
		Skipped:      0,
		LayerMapping: make(map[string]int64),
	}

	// 查找画布
	canvas, err := l.svcCtx.WorkspaceCanvasModel.FindOneByProjectId(l.ctx, uint64(req.ProjectId))
	if err != nil {
		if err == model.ErrNotFound {
			// 画布不存在，先创建画布
			createCanvasReq := &types.CreateCanvasReq{
				ProjectId:       req.ProjectId,
				Name:            "Main Canvas",
				BackgroundColor: "#ffffff",
			}
			createCanvasLogic := NewCreateCanvasLogic(l.ctx, l.svcCtx)
			_, err := createCanvasLogic.CreateCanvas(createCanvasReq)
			if err != nil {
				l.Logger.Errorf("create canvas error: %v", err)
				return nil, err
			}

			canvas, err = l.svcCtx.WorkspaceCanvasModel.FindOneByProjectId(l.ctx, uint64(req.ProjectId))
			if err != nil {
				l.Logger.Errorf("find created canvas error: %v", err)
				return nil, err
			}
		} else {
			l.Logger.Errorf("find canvas error: %v", err)
			return nil, err
		}
	}

	// 遍历所有图层进行同步
	for _, layerInput := range req.Layers {
		// Debug log
		l.Logger.Infof("Processing layer: id=%s, name=%s, type=%s, zIndex=%d",
			layerInput.Id, layerInput.Name, layerInput.LayerType, layerInput.ZIndex)

		// 序列化 properties 为 JSON
		propertiesJSON, err := json.Marshal(layerInput.Properties)
		if err != nil {
			l.Logger.Errorf("marshal properties error: %v", err)
			resp.Skipped++
			continue
		}

		l.Logger.Infof("Layer properties JSON: %s", string(propertiesJSON))

		layer := &model.WorkspaceLayer{
			CanvasId:   canvas.Id,
			LayerType:  layerInput.LayerType,
			Name:       layerInput.Name,
			ZIndex:     int32(layerInput.ZIndex),
			PositionX:  layerInput.X,
			PositionY:  layerInput.Y,
			Width:      layerInput.Width,
			Height:     layerInput.Height,
			Rotation:   layerInput.Rotation,
			Visible:    layerInput.Visible,
			Locked:     layerInput.Locked,
			Properties: string(propertiesJSON),
			Deleted:    false,
		}

		if layerInput.FileId != nil {
			layer.FileId.Valid = true
			layer.FileId.Int64 = *layerInput.FileId
		}

		// 检查是否已存在（通过本地 ID 映射，实际应该通过某种唯一标识）
		// 这里简化处理，假设所有传入的图层都是新的或需要更新的
		// 实际生产中应该通过 layerInput.Id 来查找现有图层

		// 尝试查找是否有同名图层（简化逻辑）
		existingLayers, err := l.svcCtx.WorkspaceLayerModel.FindByCanvasId(l.ctx, canvas.Id)
		if err != nil {
			l.Logger.Errorf("failed to find existing layers: %v", err)
		}
		l.Logger.Infof("Found %d existing layers for canvas %d", len(existingLayers), canvas.Id)

		found := false
		for _, existing := range existingLayers {
			l.Logger.Infof("Checking existing layer: id=%d, name=%s, type=%s",
				existing.Id, existing.Name, existing.LayerType)
			if existing.Name == layerInput.Name && existing.LayerType == layerInput.LayerType {
				// 更新现有图层
				layer.Id = existing.Id
				rowsAffected, err := l.svcCtx.WorkspaceLayerModel.Update(l.ctx, int64(layer.Id), layer)
				if err != nil {
					l.Logger.Errorf("update layer error: %v", err)
					resp.Skipped++
					continue
				}
				l.Logger.Infof("Updated layer id=%d, rowsAffected=%d", layer.Id, rowsAffected)
				resp.Updated++
				resp.LayerMapping[layerInput.Id] = int64(layer.Id)
				found = true
				break
			}
		}

		if !found {
			// 插入新图层
			layer.CreatedBy = 1 // 暂时硬编码，实际应该从 JWT 获取
			layerId, err := l.svcCtx.WorkspaceLayerModel.Insert(l.ctx, layer)
			if err != nil {
				l.Logger.Errorf("insert layer error: %v", err)
				resp.Skipped++
				continue
			}
			l.Logger.Infof("Inserted layer id=%d", layerId)
			resp.Uploaded++
			resp.LayerMapping[layerInput.Id] = layerId
		}
	}

	return
}
