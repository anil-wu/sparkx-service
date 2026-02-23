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

type CreateCanvasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCanvasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCanvasLogic {
	return &CreateCanvasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCanvasLogic) CreateCanvas(req *types.CreateCanvasReq) (resp *types.CreateCanvasResp, err error) {
	// 检查是否已经存在画布
	existingCanvas, err := l.svcCtx.WorkspaceCanvasModel.FindOneByProjectId(l.ctx, uint64(req.ProjectId))
	if err == nil && existingCanvas != nil {
		// 画布已存在，返回现有画布 ID
		return &types.CreateCanvasResp{
			CanvasId: int64(existingCanvas.Id),
		}, nil
	}

	// 创建新画布
	canvas := &model.WorkspaceCanvas{
		ProjectId:       uint64(req.ProjectId),
		Name:            req.Name,
		BackgroundColor: req.BackgroundColor,
		CreatedBy:       uint64(req.ProjectId), // 暂时使用 projectId 作为 createdBy，实际应该从 JWT 获取
	}

	// 如果有 metadata，转换为 JSON 字符串
	if req.Metadata != nil {
		// 这里需要序列化 JSON，暂时简单处理
		canvas.Metadata.Valid = true
		canvas.Metadata.String = "{}"
	}

	canvasId, err := l.svcCtx.WorkspaceCanvasModel.Insert(l.ctx, canvas)
	if err != nil {
		l.Logger.Errorf("insert canvas error: %v", err)
		return nil, err
	}

	return &types.CreateCanvasResp{
		CanvasId: canvasId,
	}, nil
}
