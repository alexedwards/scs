package memstore

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var errTypeAssertionFailed = errors.New("type assertion failed: could not convert interface{} to []byte")

func New(sweepInterval time.Duration) *MemStore {
	return &MemStore{
		cache.New(cache.DefaultExpiration, sweepInterval),
	}
}

type MemStore struct {
	*cache.Cache
}

func (m *MemStore) Find(token string) ([]byte, bool, error) {
	v, exists := m.Cache.Get(token)
	if exists == false {
		return nil, exists, nil
	}

	b, ok := v.([]byte)
	if ok == false {
		return nil, exists, errTypeAssertionFailed
	}

	return b, exists, nil
}

func (m *MemStore) Save(token string, b []byte, expiry time.Time) error {
	m.Cache.Set(token, b, expiry.Sub(time.Now()))
	return nil
}

func (m *MemStore) Delete(token string) error {
	m.Cache.Delete(token)
	return nil
}
