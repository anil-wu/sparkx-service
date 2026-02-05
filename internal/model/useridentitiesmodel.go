package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ UserIdentitiesModel = (*customUserIdentitiesModel)(nil)

type (
	// UserIdentitiesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserIdentitiesModel.
	UserIdentitiesModel interface {
		userIdentitiesModel
	}

	customUserIdentitiesModel struct {
		*defaultUserIdentitiesModel
	}
)

// NewUserIdentitiesModel returns a model for the database table.
func NewUserIdentitiesModel(conn *gorm.DB, connS sqlx.SqlConn) UserIdentitiesModel {
	return &customUserIdentitiesModel{
		defaultUserIdentitiesModel: newUserIdentitiesModel(conn, connS),
	}
}
