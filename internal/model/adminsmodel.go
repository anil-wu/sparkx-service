package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ AdminsModel = (*customAdminsModel)(nil)

type (
	AdminsModel interface {
		adminsModel
	}

	customAdminsModel struct {
		*defaultAdminsModel
	}
)

func NewAdminsModel(conn *gorm.DB, connS sqlx.SqlConn) AdminsModel {
	return &customAdminsModel{
		defaultAdminsModel: newAdminsModel(conn, connS),
	}
}

func (m *customAdminsModel) TableName() string {
	return m.defaultAdminsModel.TableName()
}
