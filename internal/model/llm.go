package model

import (
	"database/sql"
	"time"
)

type LlmProviders struct {
	Id          uint64         `db:"id" gorm:"column:id;primaryKey"`
	Name        string         `db:"name" gorm:"column:name"`
	BaseUrl     string         `db:"base_url" gorm:"column:base_url"`
	ApiKey      sql.NullString `db:"api_key" gorm:"column:api_key"`
	Description sql.NullString `db:"description" gorm:"column:description"`
	CreatedAt   time.Time      `db:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `db:"updated_at" gorm:"column:updated_at"`
}

func (LlmProviders) TableName() string { return "llm_providers" }

type LlmModels struct {
	Id               uint64    `db:"id" gorm:"column:id;primaryKey"`
	ProviderId       uint64    `db:"provider_id" gorm:"column:provider_id"`
	ModelName        string    `db:"model_name" gorm:"column:model_name"`
	ModelType        string    `db:"model_type" gorm:"column:model_type"`
	MaxInputTokens   int       `db:"max_input_tokens" gorm:"column:max_input_tokens"`
	MaxOutputTokens  int       `db:"max_output_tokens" gorm:"column:max_output_tokens"`
	SupportStream    bool      `db:"support_stream" gorm:"column:support_stream"`
	SupportJson      bool      `db:"support_json" gorm:"column:support_json"`
	PriceInputPer1k  float64   `db:"price_input_per_1k" gorm:"column:price_input_per_1k"`
	PriceOutputPer1k float64   `db:"price_output_per_1k" gorm:"column:price_output_per_1k"`
	CreatedAt        time.Time `db:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time `db:"updated_at" gorm:"column:updated_at"`
}

func (LlmModels) TableName() string { return "llm_models" }

type LlmUsageLogs struct {
	Id           uint64    `db:"id" gorm:"column:id;primaryKey"`
	LlmModelId   uint64    `db:"llm_model_id" gorm:"column:llm_model_id"`
	ProjectId    uint64    `db:"project_id" gorm:"column:project_id"`
	InputTokens  int       `db:"input_tokens" gorm:"column:input_tokens"`
	OutputTokens int       `db:"output_tokens" gorm:"column:output_tokens"`
	CacheHit     bool      `db:"cache_hit" gorm:"column:cache_hit"`
	CostUsd      float64   `db:"cost_usd" gorm:"column:cost_usd"`
	CreatedAt    time.Time `db:"created_at" gorm:"column:created_at"`
}

func (LlmUsageLogs) TableName() string { return "llm_usage_logs" }
