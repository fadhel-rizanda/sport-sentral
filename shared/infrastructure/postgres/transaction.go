package postgres

import (
	"context"
	"gorm.io/gorm"
	"microservice-golang/shared/pkg/database"
)

type txKey struct{}

// GormTxKey is the context key for the gorm.DB transaction instance.
var GormTxKey = txKey{}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) database.TransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, GormTxKey, tx)
		return fn(txCtx)
	})
}

// GetTx retrieves the gorm.DB instance from context, or returns the default db if not found.
func GetTx(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(GormTxKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
