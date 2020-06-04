package scs

import (
	"context"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func NewMock() ISession {
	return &Mock{}
}

func (m *Mock) Load(ctx context.Context, token string) (context.Context, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(context.Context), args.Error(1)
}

func (m *Mock) Commit(ctx context.Context) (string, time.Time, error) {
	args := m.Called(ctx)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *Mock) Destroy(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *Mock) Put(ctx context.Context, key string, val interface{}) {
	m.Called(ctx, key, val)
}

func (m *Mock) Get(ctx context.Context, key string) interface{} {
	return m.Called(ctx, key).Get(0).(interface{})
}

func (m *Mock) Pop(ctx context.Context, key string) interface{} {
	return m.Called(ctx, key).Get(0).(interface{})
}

func (m *Mock) Remove(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *Mock) Clear(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *Mock) Exists(ctx context.Context, key string) bool {
	return m.Called(ctx, key).Bool(0)
}

func (m *Mock) Keys(ctx context.Context) []string {
	return m.Called(ctx).Get(0).([]string)
}

func (m *Mock) RenewToken(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *Mock) Status(ctx context.Context) Status {
	return m.Called(ctx).Get(0).(Status)
}

func (m *Mock) GetString(ctx context.Context, key string) string {
	return m.Called(ctx, key).String(0)
}

func (m *Mock) GetBool(ctx context.Context, key string) bool {
	return m.Called(ctx, key).Bool(0)
}

func (m *Mock) GetInt(ctx context.Context, key string) int {
	return m.Called(ctx, key).Int(0)
}

func (m *Mock) GetFloat(ctx context.Context, key string) float64 {
	return m.Called(ctx, key).Get(0).(float64)
}

func (m *Mock) GetBytes(ctx context.Context, key string) []byte {
	return m.Called(ctx, key).Get(0).([]byte)
}

func (m *Mock) GetTime(ctx context.Context, key string) time.Time {
	return m.Called(ctx, key).Get(0).(time.Time)
}

func (m *Mock) PopString(ctx context.Context, key string) string {
	return m.Called(ctx, key).String(0)
}

func (m *Mock) PopBool(ctx context.Context, key string) bool {
	return m.Called(ctx, key).Bool(0)
}

func (m *Mock) PopInt(ctx context.Context, key string) int {
	return m.Called(ctx, key).Int(0)
}

func (m *Mock) PopFloat(ctx context.Context, key string) float64 {
	return m.Called(ctx, key).Get(0).(float64)
}

func (m *Mock) PopBytes(ctx context.Context, key string) []byte {
	return m.Called(ctx, key).Get(0).([]byte)
}

func (m *Mock) PopTime(ctx context.Context, key string) time.Time {
	return m.Called(ctx, key).Get(0).(time.Time)
}

func (m *Mock) LoadAndSave(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}
