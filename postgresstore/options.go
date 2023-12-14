package postgresstore

import (
	"time"
)

type storeOptions struct {
	sessionTableName string
	dataColumnName   string
	tokenColumnName  string
	expiryColumnName string
	cleanupInterval  time.Duration
}

type StoreOption func(*storeOptions)

func WithSessionTableName(tableName string) StoreOption {
	return func(options *storeOptions) {
		options.sessionTableName = tableName
	}
}

func WithDataColumnName(columnName string) StoreOption {
	return func(options *storeOptions) {
		options.dataColumnName = columnName
	}
}

func WithTokenColumnName(columnName string) StoreOption {
	return func(options *storeOptions) {
		options.tokenColumnName = columnName
	}
}

func WithExpiryColumnName(columnName string) StoreOption {
	return func(options *storeOptions) {
		options.expiryColumnName = columnName
	}
}

func WithCleanupInterval(interval time.Duration) StoreOption {
	return func(options *storeOptions) {
		options.cleanupInterval = interval
	}
}
