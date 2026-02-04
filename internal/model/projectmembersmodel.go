package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ ProjectMembersModel = (*customProjectMembersModel)(nil)

type (
	// ProjectMembersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProjectMembersModel.
	ProjectMembersModel interface {
		projectMembersModel
	}

	customProjectMembersModel struct {
		*defaultProjectMembersModel
	}
)

// NewProjectMembersModel returns a model for the database table.
func NewProjectMembersModel(conn *gorm.DB, connS sqlx.SqlConn) ProjectMembersModel {
	return &customProjectMembersModel{
		defaultProjectMembersModel: newProjectMembersModel(conn, connS),
	}
}
