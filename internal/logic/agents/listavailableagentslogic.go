// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

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

var agentTypes = map[string]struct{}{
	"code":   {},
	"asset":  {},
	"design": {},
	"test":   {},
	"build":  {},
	"ops":    {},
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
