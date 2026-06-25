package database

import "context"

// TransactionManager defines the interface for running operations within a transaction.
type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
