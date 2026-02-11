package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func ensureAdmin(ctx context.Context) error {
	_, ok := ctx.Value("adminId").(json.Number)
	if !ok {
		return errors.New("unauthorized")
	}
	return nil
}

type CreateAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAgentLogic {
	return &CreateAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAgentLogic) CreateAgent(req *types.CreateAgentReq) (resp *types.AgentResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	agentType := normalizeAgentType(req.AgentType)
	if agentType == "" {
		agentType = "code"
	}
	if _, ok := agentTypes[agentType]; !ok {
		return nil, errors.New("invalid agentType")
	}

	a := &model.Agents{
		Name:        strings.TrimSpace(req.Name),
		Description: sql.NullString{String: req.Description, Valid: strings.TrimSpace(req.Description) != ""},
		Instruction: sql.NullString{String: req.Instruction, Valid: strings.TrimSpace(req.Instruction) != ""},
		AgentType:   agentType,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(a).Error; err != nil {
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

type ListAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentsLogic {
	return &ListAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentsLogic) ListAgents(req *types.ListAgentsReq) (resp *types.AgentListResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
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

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Agents{})
	if t := normalizeAgentType(req.AgentType); t != "" {
		query = query.Where("`agent_type` = ?", t)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var list []model.Agents
	if err := query.Order("`id` DESC").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&list).Error; err != nil {
		return nil, err
	}

	out := make([]types.AgentResp, 0, len(list))
	for _, a := range list {
		instruction := a.Instruction.String
		if strings.TrimSpace(instruction) == "" {
			instruction = a.LegacyCommand.String
		}
		out = append(out, types.AgentResp{
			Id:          int64(a.Id),
			Name:        a.Name,
			Description: a.Description.String,
			Instruction: instruction,
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

type GetAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentLogic {
	return &GetAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentLogic) GetAgent(req *types.GetAgentReq) (resp *types.AgentResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var a model.Agents
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	instruction := a.Instruction.String
	if strings.TrimSpace(instruction) == "" {
		instruction = a.LegacyCommand.String
	}
	return &types.AgentResp{
		Id:          int64(a.Id),
		Name:        a.Name,
		Description: a.Description.String,
		Instruction: instruction,
		AgentType:   a.AgentType,
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type UpdateAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentLogic {
	return &UpdateAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentLogic) UpdateAgent(req *types.UpdateAgentReq) (resp *types.BaseResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	updates := map[string]interface{}{}
	if s := strings.TrimSpace(req.Name); s != "" {
		updates["name"] = s
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Instruction != "" {
		updates["instruction"] = req.Instruction
	}
	if req.AgentType != "" {
		agentType := normalizeAgentType(req.AgentType)
		if _, ok := agentTypes[agentType]; !ok {
			return nil, errors.New("invalid agentType")
		}
		updates["agent_type"] = agentType
	}

	if len(updates) == 0 {
		return &types.BaseResp{Code: 0, Msg: "success"}, nil
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Agents{}).Where("`id` = ?", req.Id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type DeleteAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAgentLogic {
	return &DeleteAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAgentLogic) DeleteAgent(req *types.DeleteAgentReq) (resp *types.BaseResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var cnt int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.AgentLlmBindings{}).Where("`agent_id` = ?", req.Id).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, errors.New("agent is referenced by bindings")
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).Delete(&model.Agents{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type agentBindingRow struct {
	Id           uint64    `gorm:"column:id"`
	AgentId      uint64    `gorm:"column:agent_id"`
	LlmModelId   uint64    `gorm:"column:llm_model_id"`
	Priority     int       `gorm:"column:priority"`
	IsActive     bool      `gorm:"column:is_active"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	ProviderId   uint64    `gorm:"column:provider_id"`
	ProviderName string    `gorm:"column:provider_name"`
	ModelName    string    `gorm:"column:model_name"`
	ModelType    string    `gorm:"column:model_type"`
}

func fetchAgentBinding(ctx context.Context, db *gorm.DB, id uint64) (*agentBindingRow, error) {
	var row agentBindingRow
	err := db.WithContext(ctx).
		Table("agent_llm_bindings AS b").
		Select("b.id, b.agent_id, b.llm_model_id, b.priority, b.is_active, b.created_at, m.provider_id, p.name AS provider_name, m.model_name, m.model_type").
		Joins("JOIN llm_models AS m ON m.id = b.llm_model_id").
		Joins("JOIN llm_providers AS p ON p.id = m.provider_id").
		Where("b.id = ?", id).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

type ListAgentBindingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAgentBindingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAgentBindingsLogic {
	return &ListAgentBindingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAgentBindingsLogic) ListAgentBindings(req *types.ListAgentBindingsReq) (resp *types.AgentBindingListResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.AgentId <= 0 {
		return nil, model.InputParamInvalid
	}

	var rows []agentBindingRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("agent_llm_bindings AS b").
		Select("b.id, b.agent_id, b.llm_model_id, b.priority, b.is_active, b.created_at, m.provider_id, p.name AS provider_name, m.model_name, m.model_type").
		Joins("JOIN llm_models AS m ON m.id = b.llm_model_id").
		Joins("JOIN llm_providers AS p ON p.id = m.provider_id").
		Where("b.agent_id = ?", req.AgentId).
		Order("b.priority DESC, b.id DESC").
		Limit(1).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]types.AgentBindingResp, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.AgentBindingResp{
			Id:           int64(r.Id),
			AgentId:      int64(r.AgentId),
			LlmModelId:   int64(r.LlmModelId),
			Priority:     int64(r.Priority),
			IsActive:     r.IsActive,
			CreatedAt:    r.CreatedAt.Format("2006-01-02 15:04:05"),
			ProviderId:   int64(r.ProviderId),
			ProviderName: r.ProviderName,
			ModelName:    r.ModelName,
			ModelType:    r.ModelType,
		})
	}

	return &types.AgentBindingListResp{List: out}, nil
}

type CreateAgentBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAgentBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAgentBindingLogic {
	return &CreateAgentBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAgentBindingLogic) CreateAgentBinding(req *types.CreateAgentBindingReq) (resp *types.AgentBindingResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.AgentId <= 0 || req.LlmModelId <= 0 {
		return nil, model.InputParamInvalid
	}

	var agentCnt int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Agents{}).Where("`id` = ?", req.AgentId).Count(&agentCnt).Error; err != nil {
		return nil, err
	}
	if agentCnt == 0 {
		return nil, errors.New("agent not found")
	}

	var modelCnt int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmModels{}).Where("`id` = ?", req.LlmModelId).Count(&modelCnt).Error; err != nil {
		return nil, err
	}
	if modelCnt == 0 {
		return nil, errors.New("llm model not found")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	b := &model.AgentLlmBindings{
		AgentId:    uint64(req.AgentId),
		LlmModelId: uint64(req.LlmModelId),
		Priority:   int(req.Priority),
		IsActive:   isActive,
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "agent_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"llm_model_id", "priority", "is_active"}),
		}).
		Create(b).Error; err != nil {
		return nil, err
	}

	var persisted model.AgentLlmBindings
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`agent_id` = ?", req.AgentId).First(&persisted).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	row, err := fetchAgentBinding(l.ctx, l.svcCtx.DB, persisted.Id)
	if err != nil {
		return nil, err
	}

	return &types.AgentBindingResp{
		Id:           int64(row.Id),
		AgentId:      int64(row.AgentId),
		LlmModelId:   int64(row.LlmModelId),
		Priority:     int64(row.Priority),
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Format("2006-01-02 15:04:05"),
		ProviderId:   int64(row.ProviderId),
		ProviderName: row.ProviderName,
		ModelName:    row.ModelName,
		ModelType:    row.ModelType,
	}, nil
}

type UpdateAgentBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAgentBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAgentBindingLogic {
	return &UpdateAgentBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAgentBindingLogic) UpdateAgentBinding(req *types.UpdateAgentBindingReq) (resp *types.BaseResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var existing model.AgentLlmBindings
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Priority != nil {
		updates["priority"] = int(*req.Priority)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.LlmModelId > 0 && uint64(req.LlmModelId) != existing.LlmModelId {
		var cnt int64
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmModels{}).Where("`id` = ?", req.LlmModelId).Count(&cnt).Error; err != nil {
			return nil, err
		}
		if cnt == 0 {
			return nil, errors.New("llm model not found")
		}
		updates["llm_model_id"] = req.LlmModelId
	}

	if len(updates) == 0 {
		return &types.BaseResp{Code: 0, Msg: "success"}, nil
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.AgentLlmBindings{}).Where("`id` = ?", req.Id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type DeleteAgentBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAgentBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAgentBindingLogic {
	return &DeleteAgentBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAgentBindingLogic) DeleteAgentBinding(req *types.DeleteAgentBindingReq) (resp *types.BaseResp, err error) {
	if err := ensureAdmin(l.ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).Delete(&model.AgentLlmBindings{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}
