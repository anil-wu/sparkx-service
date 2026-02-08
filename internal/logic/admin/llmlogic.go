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
)

func formatTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

func isRetryableDbErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid connection") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "connection reset")
}

func withDbRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if !isRetryableDbErr(err) || ctx.Err() != nil || i == 2 {
				return err
			}
			time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

type CreateLlmProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLlmProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLlmProviderLogic {
	return &CreateLlmProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLlmProviderLogic) CreateLlmProvider(req *types.CreateLlmProviderReq) (resp *types.LlmProviderResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	provider := &model.LlmProviders{
		Name:        req.Name,
		BaseUrl:     req.BaseUrl,
		ApiKey:      sql.NullString{String: req.ApiKey, Valid: req.ApiKey != ""},
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(provider).Error; err != nil {
		return nil, err
	}

	return &types.LlmProviderResp{
		Id:          int64(provider.Id),
		Name:        provider.Name,
		BaseUrl:     provider.BaseUrl,
		HasApiKey:   provider.ApiKey.Valid && provider.ApiKey.String != "",
		Description: provider.Description.String,
		CreatedAt:   provider.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   provider.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type UpdateLlmProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLlmProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLlmProviderLogic {
	return &UpdateLlmProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLlmProviderLogic) UpdateLlmProvider(req *types.UpdateLlmProviderReq) (resp *types.BaseResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.BaseUrl != "" {
		updates["base_url"] = req.BaseUrl
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ClearApiKey {
		updates["api_key"] = nil
	} else if req.ApiKey != "" {
		updates["api_key"] = req.ApiKey
	}

	if len(updates) == 0 {
		return &types.BaseResp{Code: 0, Msg: "success"}, nil
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmProviders{}).Where("`id` = ?", req.Id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type DeleteLlmProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLlmProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLlmProviderLogic {
	return &DeleteLlmProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLlmProviderLogic) DeleteLlmProvider(req *types.DeleteLlmProviderReq) (resp *types.BaseResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var cnt int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmModels{}).Where("`provider_id` = ?", req.Id).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, errors.New("provider is referenced by models")
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).Delete(&model.LlmProviders{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type GetLlmProviderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLlmProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLlmProviderLogic {
	return &GetLlmProviderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLlmProviderLogic) GetLlmProvider(req *types.GetLlmProviderReq) (resp *types.LlmProviderResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var provider model.LlmProviders
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).First(&provider).Error; err != nil {
		return nil, err
	}

	return &types.LlmProviderResp{
		Id:          int64(provider.Id),
		Name:        provider.Name,
		BaseUrl:     provider.BaseUrl,
		HasApiKey:   provider.ApiKey.Valid && provider.ApiKey.String != "",
		Description: provider.Description.String,
		CreatedAt:   provider.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   provider.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type ListLlmProvidersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLlmProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLlmProvidersLogic {
	return &ListLlmProvidersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLlmProvidersLogic) ListLlmProviders(req *types.PageReq) (resp *types.LlmProviderListResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmProviders{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var providers []model.LlmProviders
	offset := (page - 1) * pageSize
	if err := l.svcCtx.DB.WithContext(l.ctx).Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&providers).Error; err != nil {
		return nil, err
	}

	list := make([]types.LlmProviderResp, 0, len(providers))
	for _, p := range providers {
		list = append(list, types.LlmProviderResp{
			Id:          int64(p.Id),
			Name:        p.Name,
			BaseUrl:     p.BaseUrl,
			HasApiKey:   p.ApiKey.Valid && p.ApiKey.String != "",
			Description: p.Description.String,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.LlmProviderListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

type CreateLlmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLlmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLlmModelLogic {
	return &CreateLlmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLlmModelLogic) CreateLlmModel(req *types.CreateLlmModelReq) (resp *types.LlmModelResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.ProviderId <= 0 || req.ModelName == "" {
		return nil, model.InputParamInvalid
	}
	if req.ModelType != "llm" && req.ModelType != "vlm" && req.ModelType != "embedding" {
		return nil, model.InputParamInvalid
	}

	var provider model.LlmProviders
	if err := withDbRetry(l.ctx, func() error {
		return l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.ProviderId).First(&provider).Error
	}); err != nil {
		return nil, err
	}

	m := &model.LlmModels{
		ProviderId:       uint64(req.ProviderId),
		ModelName:        req.ModelName,
		ModelType:        req.ModelType,
		MaxInputTokens:   int(req.MaxInputTokens),
		MaxOutputTokens:  int(req.MaxOutputTokens),
		SupportStream:    req.SupportStream,
		SupportJson:      req.SupportJson,
		PriceInputPer1k:  req.PriceInputPer1k,
		PriceOutputPer1k: req.PriceOutputPer1k,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return &types.LlmModelResp{
		Id:               int64(m.Id),
		ProviderId:       int64(m.ProviderId),
		ModelName:        m.ModelName,
		ModelType:        m.ModelType,
		MaxInputTokens:   int64(m.MaxInputTokens),
		MaxOutputTokens:  int64(m.MaxOutputTokens),
		SupportStream:    m.SupportStream,
		SupportJson:      m.SupportJson,
		PriceInputPer1k:  m.PriceInputPer1k,
		PriceOutputPer1k: m.PriceOutputPer1k,
		CreatedAt:        m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type UpdateLlmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLlmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLlmModelLogic {
	return &UpdateLlmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLlmModelLogic) UpdateLlmModel(req *types.UpdateLlmModelReq) (resp *types.BaseResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	updates := map[string]interface{}{}

	if req.ProviderId != nil {
		if *req.ProviderId <= 0 {
			return nil, model.InputParamInvalid
		}
		var provider model.LlmProviders
		if err := withDbRetry(l.ctx, func() error {
			return l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", *req.ProviderId).First(&provider).Error
		}); err != nil {
			return nil, err
		}
		updates["provider_id"] = *req.ProviderId
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
	}
	if req.ModelType != "" {
		if req.ModelType != "llm" && req.ModelType != "vlm" && req.ModelType != "embedding" {
			return nil, model.InputParamInvalid
		}
		updates["model_type"] = req.ModelType
	}
	if req.MaxInputTokens != nil {
		updates["max_input_tokens"] = *req.MaxInputTokens
	}
	if req.MaxOutputTokens != nil {
		updates["max_output_tokens"] = *req.MaxOutputTokens
	}
	if req.SupportStream != nil {
		updates["support_stream"] = *req.SupportStream
	}
	if req.SupportJson != nil {
		updates["support_json"] = *req.SupportJson
	}
	if req.PriceInputPer1k != nil {
		updates["price_input_per_1k"] = *req.PriceInputPer1k
	}
	if req.PriceOutputPer1k != nil {
		updates["price_output_per_1k"] = *req.PriceOutputPer1k
	}

	if len(updates) == 0 {
		return &types.BaseResp{Code: 0, Msg: "success"}, nil
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmModels{}).Where("`id` = ?", req.Id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type DeleteLlmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLlmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLlmModelLogic {
	return &DeleteLlmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLlmModelLogic) DeleteLlmModel(req *types.DeleteLlmModelReq) (resp *types.BaseResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).Delete(&model.LlmModels{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}

	return &types.BaseResp{Code: 0, Msg: "success"}, nil
}

type GetLlmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLlmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLlmModelLogic {
	return &GetLlmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLlmModelLogic) GetLlmModel(req *types.GetLlmModelReq) (resp *types.LlmModelResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	if req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	var m model.LlmModels
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.Id).First(&m).Error; err != nil {
		return nil, err
	}

	return &types.LlmModelResp{
		Id:               int64(m.Id),
		ProviderId:       int64(m.ProviderId),
		ModelName:        m.ModelName,
		ModelType:        m.ModelType,
		MaxInputTokens:   int64(m.MaxInputTokens),
		MaxOutputTokens:  int64(m.MaxOutputTokens),
		SupportStream:    m.SupportStream,
		SupportJson:      m.SupportJson,
		PriceInputPer1k:  m.PriceInputPer1k,
		PriceOutputPer1k: m.PriceOutputPer1k,
		CreatedAt:        m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type ListLlmModelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLlmModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLlmModelsLogic {
	return &ListLlmModelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLlmModelsLogic) ListLlmModels(req *types.ListLlmModelsReq) (resp *types.LlmModelListResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmModels{})
	if req.ProviderId > 0 {
		query = query.Where("`provider_id` = ?", req.ProviderId)
	}
	if req.ModelType != "" {
		query = query.Where("`model_type` = ?", req.ModelType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var models []model.LlmModels
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&models).Error; err != nil {
		return nil, err
	}

	list := make([]types.LlmModelResp, 0, len(models))
	for _, m := range models {
		list = append(list, types.LlmModelResp{
			Id:               int64(m.Id),
			ProviderId:       int64(m.ProviderId),
			ModelName:        m.ModelName,
			ModelType:        m.ModelType,
			MaxInputTokens:   int64(m.MaxInputTokens),
			MaxOutputTokens:  int64(m.MaxOutputTokens),
			SupportStream:    m.SupportStream,
			SupportJson:      m.SupportJson,
			PriceInputPer1k:  m.PriceInputPer1k,
			PriceOutputPer1k: m.PriceOutputPer1k,
			CreatedAt:        m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.LlmModelListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

type ListLlmUsageLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLlmUsageLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLlmUsageLogsLogic {
	return &ListLlmUsageLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLlmUsageLogsLogic) ListLlmUsageLogs(req *types.ListLlmUsageLogsReq) (resp *types.LlmUsageLogListResp, err error) {
	_, ok := l.ctx.Value("adminId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.LlmUsageLogs{})
	if req.ProjectId > 0 {
		query = query.Where("`project_id` = ?", req.ProjectId)
	}
	if req.LlmModelId > 0 {
		query = query.Where("`llm_model_id` = ?", req.LlmModelId)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []model.LlmUsageLogs
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&logs).Error; err != nil {
		return nil, err
	}

	list := make([]types.LlmUsageLogResp, 0, len(logs))
	for _, it := range logs {
		list = append(list, types.LlmUsageLogResp{
			Id:           int64(it.Id),
			LlmModelId:   int64(it.LlmModelId),
			ProjectId:    int64(it.ProjectId),
			InputTokens:  int64(it.InputTokens),
			OutputTokens: int64(it.OutputTokens),
			CacheHit:     it.CacheHit,
			CostUsd:      it.CostUsd,
			CreatedAt:    it.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.LlmUsageLogListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}
