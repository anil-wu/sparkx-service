package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ FileVersionsModel = (*customFileVersionsModel)(nil)

type (
	// FileVersionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFileVersionsModel.
	FileVersionsModel interface {
		fileVersionsModel
	}

	customFileVersionsModel struct {
		*defaultFileVersionsModel
	}
)

// NewFileVersionsModel returns a model for the database table.
func NewFileVersionsModel(conn *gorm.DB, connS sqlx.SqlConn) FileVersionsModel {
	return &customFileVersionsModel{
		defaultFileVersionsModel: newFileVersionsModel(conn, connS),
	}
}
