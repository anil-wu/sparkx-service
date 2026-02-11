// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

var agentTypes = map[string]struct{}{
	"code":    {},
	"asset":   {},
	"design":  {},
	"test":    {},
	"build":   {},
	"ops":     {},
	"project": {},
}

func normalizeAgentType(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func ensureUser(ctx context.Context) error {
	_, ok := ctx.Value("userId").(json.Number)
	if !ok {
		return errors.New("unauthorized")
	}
	return nil
}

type ListAvailableAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAvailableAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAvailableAgentsLogic {
	return &ListAvailableAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAvailableAgentsLogic) ListAvailableAgents(req *types.ListAgentsReq) (resp *types.AgentListResp, err error) {
	if err := ensureUser(l.ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, model.InputParamInvalid
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Agents{}).
		Joins("JOIN agent_llm_bindings AS b ON b.agent_id = agents.id AND b.is_active = ?", true)

	if t := normalizeAgentType(req.AgentType); t != "" {
		if _, ok := agentTypes[t]; !ok {
			return nil, errors.New("invalid agentType")
		}
		query = query.Where("agents.agent_type = ?", t)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []model.Agents
	if err := query.
		Select("agents.*").
		Order("agents.id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&list).Error; err != nil {
		return nil, err
	}

	out := make([]types.AgentResp, 0, len(list))
	for _, a := range list {
		out = append(out, types.AgentResp{
			Id:          int64(a.Id),
			Name:        a.Name,
			Description: a.Description.String,
			Instruction: a.Instruction.String,
			AgentType:   a.AgentType,
			CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.AgentListResp{
		List: out,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

type ListAgentConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentConfigsLogic {
	return &ListAgentConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentConfigsLogic) ListAgentConfigs(req *types.ListAgentConfigsReq) (resp *types.AgentConfigListResp, err error) {
	if err := ensureUser(l.ctx); err != nil {
		return nil, err
	}

	query := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Agents{}).
		Joins("JOIN agent_llm_bindings AS b ON b.agent_id = agents.id AND b.is_active = ?", true)

	if req != nil {
		if t := normalizeAgentType(req.AgentType); t != "" {
			if _, ok := agentTypes[t]; !ok {
				return nil, errors.New("invalid agentType")
			}
			query = query.Where("agents.agent_type = ?", t)
		}
	}

	var agents []model.Agents
	if err := query.Select("agents.*").Order("agents.id DESC").Find(&agents).Error; err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return &types.AgentConfigListResp{
			Models:     []types.AgentModelResp{},
			Agentinfos: []types.AgentInfoResp{},
		}, nil
	}

	agentIds := make([]uint64, 0, len(agents))
	for _, a := range agents {
		agentIds = append(agentIds, a.Id)
	}

	type agentBindingRowFull struct {
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

	var rows []agentBindingRowFull
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("agent_llm_bindings AS b").
		Select("b.id, b.agent_id, b.llm_model_id, b.priority, b.is_active, b.created_at, m.provider_id, p.name AS provider_name, p.base_url AS provider_base_url, p.api_key AS provider_api_key, (p.api_key IS NOT NULL AND p.api_key <> '') AS provider_has_api_key, m.model_name, m.model_type").
		Joins("JOIN llm_models AS m ON m.id = b.llm_model_id").
		Joins("JOIN llm_providers AS p ON p.id = m.provider_id").
		Where("b.agent_id IN ? AND b.is_active = ?", agentIds, true).
		Order("b.priority DESC, b.id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	bindingsByAgent := make(map[uint64][]types.AgentInfoBindingResp, len(agents))
	modelIndexById := make(map[uint64]int, len(rows))
	models := make([]types.AgentModelResp, 0, len(rows))

	for _, r := range rows {
		idx, ok := modelIndexById[r.LlmModelId]
		if !ok {
			idx = len(models)
			modelIndexById[r.LlmModelId] = idx
			models = append(models, types.AgentModelResp{
				LlmModelId:        int64(r.LlmModelId),
				ProviderId:        int64(r.ProviderId),
				ProviderName:      r.ProviderName,
				ProviderBaseUrl:   r.ProviderBaseUrl,
				ProviderApiKey:    "",
				ProviderHasApiKey: r.ProviderHasApiKey,
				ModelName:         r.ModelName,
				ModelType:         r.ModelType,
			})
		}

		bindingsByAgent[r.AgentId] = append(bindingsByAgent[r.AgentId], types.AgentInfoBindingResp{
			Id:         int64(r.Id),
			Priority:   int64(r.Priority),
			IsActive:   r.IsActive,
			ModelIndex: int64(idx),
		})
	}

	out := make([]types.AgentInfoResp, 0, len(agents))
	for _, a := range agents {
		out = append(out, types.AgentInfoResp{
			Agent: types.AgentResp{
				Id:          int64(a.Id),
				Name:        a.Name,
				Description: a.Description.String,
				Instruction: a.Instruction.String,
				AgentType:   a.AgentType,
				CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
			},
			Bindings: bindingsByAgent[a.Id],
		})
	}

	return &types.AgentConfigListResp{
		Models:     models,
		Agentinfos: out,
	}, nil
}
