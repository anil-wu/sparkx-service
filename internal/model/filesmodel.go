package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ FilesModel = (*customFilesModel)(nil)

type (
	// FilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFilesModel.
	FilesModel interface {
		filesModel
	}

	customFilesModel struct {
		*defaultFilesModel
	}
)

// NewFilesModel returns a model for the database table.
func NewFilesModel(conn *gorm.DB, connS sqlx.SqlConn) FilesModel {
	return &customFilesModel{
		defaultFilesModel: newFilesModel(conn, connS),
	}
}
