package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ ProjectsModel = (*customProjectsModel)(nil)

type (
	// ProjectsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProjectsModel.
	ProjectsModel interface {
		projectsModel
	}

	customProjectsModel struct {
		*defaultProjectsModel
	}
)

// NewProjectsModel returns a model for the database table.
func NewProjectsModel(conn *gorm.DB, connS sqlx.SqlConn) ProjectsModel {
	return &customProjectsModel{
		defaultProjectsModel: newProjectsModel(conn, connS),
	}
}
