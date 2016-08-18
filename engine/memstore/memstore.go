package memstore

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed: could not convert interface{} to []byte")

func New() *memstore {
	return &memstore{
		// Clear up expired items once every minute
		cache.New(cache.DefaultExpiration, time.Minute),
	}
}

type memstore struct {
	*cache.Cache
}

func (m *memstore) Find(token string) ([]byte, bool, error) {
	v, exists := m.Cache.Get(token)
	if exists == false {
		return nil, exists, nil
	}

	b, ok := v.([]byte)
	if ok == false {
		return nil, exists, ErrTypeAssertionFailed
	}

	return b, exists, nil
}

func (m *memstore) Save(token string, b []byte, expiry time.Time) error {
	m.Cache.Set(token, b, expiry.Sub(time.Now()))
	return nil
}

func (m *memstore) Delete(token string) error {
	m.Cache.Delete(token)
	return nil
}
