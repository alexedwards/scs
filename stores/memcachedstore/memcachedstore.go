package memcachedstore

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// Prefix controls the Memcached key prefix. You should only need to change this if there is
// a naming clash.
var Prefix = "scs:session:"

type MemcachedStore struct {
	client *memcache.Client
}

// New returns a new MemcachedStore instance.
// The conn parameter should be a pointer to a gomemcache connection pool.
func New(client *memcache.Client) *MemcachedStore {
	return &MemcachedStore{client}
}

// Find return the data for a session token from the MemcachedStore instance.
// If the session token is not found or is expired, the found return value
// is false (and the err return value is nil).
func (m *MemcachedStore) Find(token string) (b []byte, found bool, err error) {
	item, err := m.client.Get(Prefix + token)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, false, nil
		}
		return nil, false, err
	}

	return item.Value, true, nil
}

// Save adds a session token and data to the MemcachedStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (m *MemcachedStore) Save(token string, b []byte, expiry time.Time) error {
	return m.client.Set(&memcache.Item{
		Key:        Prefix + token,
		Value:      b,
		Expiration: createOffset(expiry),
	})
}

// Delete removes a session token and corresponding data from the MemcachedStore instance.
func (m *MemcachedStore) Delete(token string) error {
	return m.client.Delete(Prefix + token)
}

// createOffset calculates how expiration dates should be stored
// Memcached stores dates either as seconds since the Unix epoch OR as a relative offset from now
// It decides this by whether the offset is greater than the number of seconds in 30 days
func createOffset(expiry time.Time) int32 {
	if expiry.After(time.Now().AddDate(0, 0, 30)) { // more than 30 days away
		return int32(expiry.Unix()) // uh oh! https://en.wikipedia.org/wiki/Year_2038_problem
	}

	return int32(time.Until(expiry).Seconds())
}
