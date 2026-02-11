// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

import (
	"context"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAvailableAgentBindingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAvailableAgentBindingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableAgentBindingsLogic {
	return &ListAvailableAgentBindingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAvailableAgentBindingsLogic) ListAvailableAgentBindings(req *types.ListAgentBindingsReq) (resp *types.AgentBindingListResp, err error) {
	if err := ensureUser(l.ctx); err != nil {
		return nil, err
	}
	if req == nil || req.AgentId <= 0 {
		return nil, model.InputParamInvalid
	}

	var agentCnt int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Agents{}).Where("id = ?", req.AgentId).Count(&agentCnt).Error; err != nil {
		return nil, err
	}
	if agentCnt == 0 {
		return nil, model.ErrNotFound
	}

	type agentBindingRow struct {
		Id                uint64    `gorm:"column:id"`
		AgentId           uint64    `gorm:"column:agent_id"`
		LlmModelId        uint64    `gorm:"column:llm_model_id"`
		Priority          int       `gorm:"column:priority"`
		IsActive          bool      `gorm:"column:is_active"`
		CreatedAt         time.Time `gorm:"column:created_at"`
		ProviderId        uint64    `gorm:"column:provider_id"`
		ProviderName      string    `gorm:"column:provider_name"`
		ProviderBaseUrl   string    `gorm:"column:provider_base_url"`
		ProviderApiKey    string    `gorm:"column:provider_api_key"`
		ProviderHasApiKey bool      `gorm:"column:provider_has_api_key"`
		ModelName         string    `gorm:"column:model_name"`
		ModelType         string    `gorm:"column:model_type"`
	}

	var rows []agentBindingRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("agent_llm_bindings AS b").
		Select("b.id, b.agent_id, b.llm_model_id, b.priority, b.is_active, b.created_at, m.provider_id, p.name AS provider_name, p.base_url AS provider_base_url, p.api_key AS provider_api_key, (p.api_key IS NOT NULL AND p.api_key <> '') AS provider_has_api_key, m.model_name, m.model_type").
		Joins("JOIN llm_models AS m ON m.id = b.llm_model_id").
		Joins("JOIN llm_providers AS p ON p.id = m.provider_id").
		Where("b.agent_id = ? AND b.is_active = ?", req.AgentId, true).
		Order("b.priority DESC, b.id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]types.AgentBindingResp, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.AgentBindingResp{
			Id:                int64(r.Id),
			AgentId:           int64(r.AgentId),
			LlmModelId:        int64(r.LlmModelId),
			Priority:          int64(r.Priority),
			IsActive:          r.IsActive,
			CreatedAt:         r.CreatedAt.Format("2006-01-02 15:04:05"),
			ProviderId:        int64(r.ProviderId),
			ProviderName:      r.ProviderName,
			ProviderBaseUrl:   r.ProviderBaseUrl,
			ProviderApiKey:    r.ProviderApiKey,
			ProviderHasApiKey: r.ProviderHasApiKey,
			ModelName:         r.ModelName,
			ModelType:         r.ModelType,
		})
	}

	return &types.AgentBindingListResp{List: out}, nil
}
