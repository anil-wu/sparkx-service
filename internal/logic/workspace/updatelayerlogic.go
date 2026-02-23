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

type UpdateLayerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLayerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLayerLogic {
	return &UpdateLayerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLayerLogic) UpdateLayer(req *types.UpdateLayerReq) (resp *types.BaseResp, err error) {
	// 查找图层
	layer, err := l.svcCtx.WorkspaceLayerModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		if err == model.ErrNotFound {
			l.Logger.Errorf("layer not found: %d", req.Id)
			return nil, err
		}
		l.Logger.Errorf("find layer error: %v", err)
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		layer.Name = *req.Name
	}
	if req.X != nil {
		layer.PositionX = *req.X
	}
	if req.Y != nil {
		layer.PositionY = *req.Y
	}
	if req.Width != nil {
		layer.Width = *req.Width
	}
	if req.Height != nil {
		layer.Height = *req.Height
	}
	if req.Rotation != nil {
		layer.Rotation = *req.Rotation
	}
	if req.ZIndex != nil {
		layer.ZIndex = int32(*req.ZIndex)
	}
	if req.Visible != nil {
		layer.Visible = *req.Visible
	}
	if req.Locked != nil {
		layer.Locked = *req.Locked
	}
	if req.Properties != nil {
		propertiesJSON, err := json.Marshal(req.Properties)
		if err != nil {
			l.Logger.Errorf("marshal properties error: %v", err)
			return nil, err
		}
		layer.Properties = string(propertiesJSON)
	}

	// 执行更新
	_, err = l.svcCtx.WorkspaceLayerModel.Update(l.ctx, req.Id, layer)
	if err != nil {
		l.Logger.Errorf("update layer error: %v", err)
		return nil, err
	}

	return &types.BaseResp{
		Code: 200,
		Msg:  "success",
	}, nil
}
