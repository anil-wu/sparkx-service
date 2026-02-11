// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

import (
	"context"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetAvailableAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAvailableAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailableAgentLogic {
	return &GetAvailableAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvailableAgentLogic) GetAvailableAgent(req *types.GetAgentReq) (resp *types.AgentResp, err error) {
	if err := ensureUser(l.ctx); err != nil {
		return nil, err
	}
	if req == nil || req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var a model.Agents
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Agents{}).
		Joins("JOIN agent_llm_bindings AS b ON b.agent_id = agents.id AND b.is_active = ?", true).
		Where("agents.id = ?", req.Id).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &types.AgentResp{
		Id:          int64(a.Id),
		Name:        a.Name,
		Description: a.Description.String,
		Instruction: a.Instruction.String,
		AgentType:   a.AgentType,
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
