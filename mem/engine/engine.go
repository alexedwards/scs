package engine

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed: could not convert interface{} to []byte")

func New() *engine {
	return &engine{
		// Clear up expired items once every minute
		cache.New(cache.DefaultExpiration, time.Minute),
	}
}

type engine struct {
	*cache.Cache
}

func (e *engine) FindValues(token string) ([]byte, bool, error) {
	v, exists := e.Cache.Get(token)
	if exists == false {
		return nil, exists, nil
	}

	b, ok := v.([]byte)
	if ok == false {
		return nil, exists, ErrTypeAssertionFailed
	}

	return b, exists, nil
}

func (e *engine) Save(token string, b []byte, lifetime time.Duration) error {
	e.Cache.Set(token, b, lifetime)
	return nil
}

func (e *engine) Delete(token string) error {
	e.Cache.Delete(token)
	return nil
}
