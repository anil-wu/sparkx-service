package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ ProjectFilesModel = (*customProjectFilesModel)(nil)

type (
	// ProjectFilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProjectFilesModel.
	ProjectFilesModel interface {
		projectFilesModel
	}

	customProjectFilesModel struct {
		*defaultProjectFilesModel
	}
)

// NewProjectFilesModel returns a model for the database table.
func NewProjectFilesModel(conn *gorm.DB, connS sqlx.SqlConn) ProjectFilesModel {
	return &customProjectFilesModel{
		defaultProjectFilesModel: newProjectFilesModel(conn, connS),
	}
}
