// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetAgentConfigByNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAgentConfigByNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentConfigByNameLogic {
	return &GetAgentConfigByNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentConfigByNameLogic) GetAgentConfigByName(req *types.GetAgentByNameReq) (resp *types.AgentConfigResp, err error) {
	if err := ensureUser(l.ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, model.InputParamInvalid
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, model.InputParamInvalid
	}

	var a model.Agents
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Agents{}).
		Joins("JOIN agent_llm_bindings AS b ON b.agent_id = agents.id AND b.is_active = ?", true).
		Where("agents.name = ?", name).
		Order("agents.id DESC").
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
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
		ProviderHasApiKey bool      `gorm:"column:provider_has_api_key"`
		ModelName         string    `gorm:"column:model_name"`
		ModelType         string    `gorm:"column:model_type"`
	}

	var rows []agentBindingRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("agent_llm_bindings AS b").
		Select("b.id, b.agent_id, b.llm_model_id, b.priority, b.is_active, b.created_at, m.provider_id, p.name AS provider_name, p.base_url AS provider_base_url, (p.api_key IS NOT NULL AND p.api_key <> '') AS provider_has_api_key, m.model_name, m.model_type").
		Joins("JOIN llm_models AS m ON m.id = b.llm_model_id").
		Joins("JOIN llm_providers AS p ON p.id = m.provider_id").
		Where("b.agent_id = ? AND b.is_active = ?", a.Id, true).
		Order("b.priority DESC, b.id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	bindings := make([]types.AgentBindingResp, 0, len(rows))
	for _, r := range rows {
		bindings = append(bindings, types.AgentBindingResp{
			Id:                int64(r.Id),
			AgentId:           int64(r.AgentId),
			LlmModelId:        int64(r.LlmModelId),
			Priority:          int64(r.Priority),
			IsActive:          r.IsActive,
			CreatedAt:         r.CreatedAt.Format("2006-01-02 15:04:05"),
			ProviderId:        int64(r.ProviderId),
			ProviderName:      r.ProviderName,
			ProviderBaseUrl:   r.ProviderBaseUrl,
			ProviderHasApiKey: r.ProviderHasApiKey,
			ModelName:         r.ModelName,
			ModelType:         r.ModelType,
		})
	}

	instruction := a.Instruction.String
	if strings.TrimSpace(instruction) == "" {
		instruction = a.LegacyCommand.String
	}
	return &types.AgentConfigResp{
		Agent: types.AgentResp{
			Id:          int64(a.Id),
			Name:        a.Name,
			Description: a.Description.String,
			Instruction: instruction,
			AgentType:   a.AgentType,
			CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		Bindings: bindings,
	}, nil
}
