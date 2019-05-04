package redisstore

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// RedisStore represents the session store.
type RedisStore struct {
	pool   *redis.Pool
	prefix string
}

// New returns a new RedisStore instance. The pool parameter should be a pointer
// to a redigo connection pool. See https://godoc.org/github.com/gomodule/redigo/redis#Pool.
func New(pool *redis.Pool) *RedisStore {
	return NewWithPrefix(pool, "scs:session:")
}

// NewWithPrefix returns a new RedisStore instance. The pool parameter should be a pointer
// to a redigo connection pool. The prefix parameter controls the Redis key
// prefix, which can be used to avoid naming clashes if necessary.
func NewWithPrefix(pool *redis.Pool, prefix string) *RedisStore {
	return &RedisStore{
		pool:   pool,
		prefix: prefix,
	}
}

// Find returns the data for a given session token from the RedisStore instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.
func (r *RedisStore) Find(token string) (b []byte, exists bool, err error) {
	conn := r.pool.Get()
	defer conn.Close()

	b, err = redis.Bytes(conn.Do("GET", r.prefix+token))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the RedisStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (r *RedisStore) Commit(token string, b []byte, expiry time.Time) error {
	conn := r.pool.Get()
	defer conn.Close()

	err := conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("SET", r.prefix+token, b)
	if err != nil {
		return err
	}
	err = conn.Send("PEXPIREAT", r.prefix+token, makeMillisecondTimestamp(expiry))
	if err != nil {
		return err
	}
	_, err = conn.Do("EXEC")
	return err
}

// Delete removes a session token and corresponding data from the RedisStore
// instance.
func (r *RedisStore) Delete(token string) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", r.prefix+token)
	return err
}

func makeMillisecondTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
