package scs

import (
	"context"
	"net/http"
	"time"
)

type MockableSession interface {
	LoadAndSave(next http.Handler) http.Handler
	Load(ctx context.Context, token string) (context.Context, error)
	Commit(ctx context.Context) (string, time.Time, error)
	Destroy(ctx context.Context) error
	Put(ctx context.Context, key string, val interface{})
	Get(ctx context.Context, key string) interface{}
	Pop(ctx context.Context, key string) interface{}
	Remove(ctx context.Context, key string)
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) bool
	Keys(ctx context.Context) []string
	RenewToken(ctx context.Context) error
	Status(ctx context.Context) Status
	GetString(ctx context.Context, key string) string
	GetBool(ctx context.Context, key string) bool
	GetInt(ctx context.Context, key string) int
	GetFloat(ctx context.Context, key string) float64
	GetBytes(ctx context.Context, key string) []byte
	GetTime(ctx context.Context, key string) time.Time
	PopString(ctx context.Context, key string) string
	PopBool(ctx context.Context, key string) bool
	PopInt(ctx context.Context, key string) int
	PopFloat(ctx context.Context, key string) float64
	PopBytes(ctx context.Context, key string) []byte
	PopTime(ctx context.Context, key string) time.Time
}

// Allows to cast SessionManager to MockableSession interface
func (s *SessionManager) AsInterface() MockableSession {
	return s
}
