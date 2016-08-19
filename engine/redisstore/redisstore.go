package redisstore

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var Prefix = "scs:session:"

type RedisStore struct {
	pool *redis.Pool
}

func New(pool *redis.Pool) *RedisStore {
	return &RedisStore{pool}
}

func (r *RedisStore) Find(token string) ([]byte, bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", Prefix+token))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (r *RedisStore) Save(token string, b []byte, expiry time.Time) error {
	conn := r.pool.Get()
	defer conn.Close()

	err := conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("SET", Prefix+token, b)
	if err != nil {
		return err
	}
	err = conn.Send("PEXPIREAT", Prefix+token, makeMillisecondTimestamp(expiry))
	if err != nil {
		return err
	}
	_, err = conn.Do("EXEC")
	return err
}

func (r *RedisStore) Delete(token string) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", Prefix+"session_token")
	return err
}

func makeMillisecondTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
