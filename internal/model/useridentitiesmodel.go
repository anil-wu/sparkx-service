package model

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"time"

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

func (m *customUserIdentitiesModel) FindOneByProviderProviderUid(
	ctx context.Context,
	provider string,
	providerUid string,
) (*UserIdentities, error) {
	result, err := m.defaultUserIdentitiesModel.FindOneByProviderProviderUid(ctx, provider, providerUid)
	if err == nil || errors.Is(err, ErrNotFound) {
		return result, err
	}

	if errors.Is(err, driver.ErrBadConn) || strings.Contains(err.Error(), "invalid connection") {
		if sqlDB, e := m.DB.DB(); e == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			_ = sqlDB.PingContext(pingCtx)
			cancel()
		}
		return m.defaultUserIdentitiesModel.FindOneByProviderProviderUid(ctx, provider, providerUid)
	}

	return result, err
}
