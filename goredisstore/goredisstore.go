package goredisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore represents the session store.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// New returns a new RedisStore instance. The client parameter should be a pointer
// to a go-redis connection.
func New(client *redis.Client) *RedisStore {
	return NewWithPrefix(client, "scs:session:")
}

// NewWithPrefix returns a new RedisStore instance. The pool parameter should be a pointer
// to a redigo connection pool. The prefix parameter controls the Redis key
// prefix, which can be used to avoid naming clashes if necessary.
func NewWithPrefix(client *redis.Client, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// FindCtx returns the data for a given session token from the RedisStore instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.
func (r *RedisStore) FindCtx(ctx context.Context, token string) (b []byte, exists bool, err error) {
	b, err = r.client.Get(ctx, r.prefix+token).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// CommitCtx adds a session token and data to the RedisStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (r *RedisStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	err := r.client.Set(ctx, r.prefix+token, string(b), expiry.Sub(time.Now())).Err()
	return err
}

// DeleteCtx removes a session token and corresponding data from the RedisStore
// instance.
func (r *RedisStore) DeleteCtx(ctx context.Context, token string) error {
	return r.client.Del(ctx, r.prefix+token).Err()
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the RedisStore instance.
func (r *RedisStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	var cursor uint64
	sessions := make(map[string][]byte)

	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, r.prefix+"*", 0).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, nil
			} else {
				return nil, err
			}
		}
		for _, key := range keys {
			token := key[len(r.prefix):]
			data, exists, err := r.FindCtx(ctx, token)
			if err != nil {
				return nil, err
			}
			if exists {
				sessions[token] = data
			}
		}
		if cursor == 0 {
			break
		}
	}
	return sessions, nil
}

//
// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a seperate type makes any errors caught
// only at runtime instead of compile time. Oh well.

func (r *RedisStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}
func (r *RedisStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}
func (r *RedisStore) Delete(token string) error {
	panic("missing context arg")
}
