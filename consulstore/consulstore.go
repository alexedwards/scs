package consulstore

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulStore represents the session store.
type ConsulStore struct {
	client      *api.Client
	kv          *api.KV
	prefix      string
	stopCleanup chan bool
}

// New returns a new ConsulStore instance.
// The client parameter should be a pointer to a Consul client instance.
func New(client *api.Client) *ConsulStore {
	return NewWithOptions(client, time.Minute, "scs:session:")
}

// NewWithOptions returns a new ConsulStore instance. The client parameter should be a pointer
// to a Consul client instance. The prefix parameter controls the Consul key
// prefix, which can be used to avoid naming clashes if necessary. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithOptions(client *api.Client, cleanupInterval time.Duration, prefix string) *ConsulStore {
	c := &ConsulStore{
		client: client,
		kv:     client.KV(),
		prefix: prefix,
	}

	if cleanupInterval > 0 {
		go c.startCleanup(cleanupInterval)
	}

	return c
}

// Find returns the data for a given session token from the ConsulStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (c *ConsulStore) Find(token string) (b []byte, exists bool, err error) {
	pair, _, err := c.kv.Get(c.prefix+token, nil)
	if err != nil {
		return nil, false, err
	}

	if pair == nil {
		return nil, false, nil
	}

	if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(pair.Value[:8]) {
		return nil, false, nil
	}

	return pair.Value[8:], true, nil
}

// Commit adds a session token and data to the ConsulStore instance with the
// given expiry time. If the session token already exists then the data and expiry
// time are updated.
func (c *ConsulStore) Commit(token string, b []byte, expiry time.Time) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(expiry.UnixNano()))
	val := append(buf, b...)

	pair := &api.KVPair{Key: c.prefix + token, Value: val}
	_, err := c.kv.Put(pair, nil)
	return err
}

// Delete removes a session token and corresponding data from the ConsulStore
// instance.
func (c *ConsulStore) Delete(token string) error {
	_, err := c.kv.Delete(c.prefix+token, nil)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the ConsulStore instance.
func (c *ConsulStore) All() (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	pairs, _, err := c.kv.List(c.prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		if binary.BigEndian.Uint64(pair.Value[:8]) > uint64(time.Now().UnixNano()) {
			sessions[string(pair.Key)[len(c.prefix):]] = pair.Value[8:]
		}
	}

	return sessions, nil
}

func (c *ConsulStore) startCleanup(cleanupInterval time.Duration) {
	c.stopCleanup = make(chan bool)
	ticker := time.NewTicker(cleanupInterval)
	for {
		select {
		case <-ticker.C:
			err := c.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-c.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the ConsulStore
// instance. It's rare to terminate this; generally ConsulStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the ConsulStore is transient.
// An example is creating a new ConsulStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// ConsulStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (c *ConsulStore) StopCleanup() {
	if c.stopCleanup != nil {
		c.stopCleanup <- true
	}
}

func (c *ConsulStore) deleteExpired() error {
	pairs, _, err := c.kv.List(c.prefix, nil)
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(pair.Value[:8]) {
			c.kv.Delete(pair.Key, nil)
		}
	}

	return nil
}
