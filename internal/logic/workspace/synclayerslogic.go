// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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
	userId, err := ensureProjectMember(l.ctx, l.svcCtx, req.ProjectId)
	if err != nil {
		return nil, err
	}

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

	// 步骤 1: 先将画布的所有现有图层标记为已删除
	existingLayers, err := l.svcCtx.WorkspaceLayerModel.FindByCanvasId(l.ctx, canvas.Id)
	if err != nil {
		l.Logger.Errorf("failed to find existing layers: %v", err)
		return nil, err
	}

	l.Logger.Infof("Found %d existing layers for canvas %d, marking all as deleted", len(existingLayers), canvas.Id)

	for _, existing := range existingLayers {
		_, err := l.svcCtx.WorkspaceLayerModel.SoftDelete(l.ctx, existing.Id, uint64(userId))
		if err != nil {
			l.Logger.Errorf("soft delete existing layer %d error: %v", existing.Id, err)
			return nil, err
		}
	}

	// 步骤 2: 遍历所有需要同步的图层，恢复或创建新图层
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

		// 插入新图层（因为所有旧图层已被标记为删除）
		layer.CreatedBy = uint64(userId)
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

	return
}

func parseBoolContextValue(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		s := strings.TrimSpace(strings.ToLower(v))
		if s == "true" || s == "1" || s == "yes" {
			return true, true
		}
		if s == "false" || s == "0" || s == "no" {
			return false, true
		}
		return false, false
	case float64:
		return v != 0, true
	case int64:
		return v != 0, true
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return false, false
		}
		return n != 0, true
	default:
		return false, false
	}
}

func isSuperFromContext(ctx context.Context) bool {
	v, ok := parseBoolContextValue(ctx.Value("isSuper"))
	return ok && v
}

func ensureProjectMember(ctx context.Context, svcCtx *svc.ServiceContext, projectId int64) (int64, error) {
	userIdNumber, ok := ctx.Value("userId").(json.Number)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()
	if userId <= 0 {
		return 0, errors.New("unauthorized")
	}
	if isSuperFromContext(ctx) {
		return userId, nil
	}
	if projectId <= 0 {
		return 0, model.InputParamInvalid
	}
	var count int64
	if err := svcCtx.DB.WithContext(ctx).Model(&model.ProjectMembers{}).Where("project_id = ? AND user_id = ?", projectId, userId).Count(&count).Error; err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, errors.New("project not found or permission denied")
	}
	return userId, nil
}
