package model

import (
	"database/sql"
	"time"
)

type Agents struct {
	Id          uint64         `db:"id" gorm:"column:id;primaryKey"`
	Name        string         `db:"name" gorm:"column:name"`
	Description sql.NullString `db:"description" gorm:"column:description"`
	Instruction sql.NullString `db:"instruction" gorm:"column:instruction"`
	AgentType   string         `db:"agent_type" gorm:"column:agent_type"`
	CreatedAt   time.Time      `db:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time      `db:"updated_at" gorm:"column:updated_at"`
}

func (Agents) TableName() string { return "agents" }

type AgentLlmBindings struct {
	Id         uint64    `db:"id" gorm:"column:id;primaryKey"`
	AgentId    uint64    `db:"agent_id" gorm:"column:agent_id"`
	LlmModelId uint64    `db:"llm_model_id" gorm:"column:llm_model_id"`
	Priority   int       `db:"priority" gorm:"column:priority"`
	IsActive   bool      `db:"is_active" gorm:"column:is_active"`
	CreatedAt  time.Time `db:"created_at" gorm:"column:created_at"`
}

func (AgentLlmBindings) TableName() string { return "agent_llm_bindings" }
