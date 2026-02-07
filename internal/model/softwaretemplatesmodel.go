package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ SoftwareTemplatesModel = (*customSoftwareTemplatesModel)(nil)

type (
	// SoftwareTemplatesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSoftwareTemplatesModel.
	SoftwareTemplatesModel interface {
		softwareTemplatesModel
	}

	customSoftwareTemplatesModel struct {
		*defaultSoftwareTemplatesModel
	}
)

// NewSoftwareTemplatesModel returns a model for the database table.
func NewSoftwareTemplatesModel(conn *gorm.DB, connS sqlx.SqlConn) SoftwareTemplatesModel {
	return &customSoftwareTemplatesModel{
		defaultSoftwareTemplatesModel: newSoftwareTemplatesModel(conn, connS),
	}
}
