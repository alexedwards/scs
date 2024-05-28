package natsstore

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go/encoders/builtin"
	"github.com/nats-io/nats.go/jetstream"
)

var encoder = &builtin.GobEncoder{}

type expirableValue struct {
	Value   []byte
	Expires time.Time
}

type NatsStore struct {
	db jetstream.KeyValue

	timeout     time.Duration
	cleanup     time.Duration
	stopCleanup chan bool
}

func (ns *NatsStore) get(ctx context.Context, key string, now time.Time) ([]byte, bool, error) {
	val, err := ns.db.Get(ctx, key)
	if err == jetstream.ErrKeyNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var decoded expirableValue
	err = encoder.Decode(key, val.Value(), &decoded)
	if err != nil {
		return nil, false, err
	}

	if decoded.Expires.Before(now) {
		err = ns.delete(ctx, key)
		return nil, false, err
	}

	return decoded.Value, true, nil
}

func (ns *NatsStore) put(ctx context.Context, key string, val []byte, expiry time.Time) error {
	toEncode := expirableValue{
		Value:   val,
		Expires: expiry,
	}
	encoded, err := encoder.Encode(key, toEncode)
	if err != nil {
		return err
	}
	_, err = ns.db.Put(ctx, key, encoded)
	return err
}

func (ns *NatsStore) delete(ctx context.Context, key string) error {
	return ns.db.Purge(ctx, key)
}

// AllCtx implements scs.IterableCtxStore.
func (ns *NatsStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	// cleanup := false
	now := time.Now()

	keys, err := ns.db.ListKeys(ctx, jetstream.IgnoreDeletes())
	defer keys.Stop()
	if err != nil {
		return nil, err
	}

	out := make(map[string][]byte)
	for key := range keys.Keys() {
		val, available, err := ns.get(ctx, key, now)
		if !available || err != nil {
			// cleanup = true
			continue
		}
		out[key] = val
	}
	// ns.db.PurgeDeletes(ctx)
	return out, nil
}

func (ns *NatsStore) StartCleanup() {
	ns.stopCleanup = make(chan bool)
	ticker := time.NewTicker(ns.cleanup)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), ns.cleanup.Truncate(time.Second))

			err := ns.db.PurgeDeletes(ctx)
			if err != nil {
				log.Println(err)
			}
			cancel()
		case <-ns.stopCleanup:
			ns.stopCleanup = nil
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the NetsKVStore
// instance. It's rare to terminate this; generally NetsKVStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the NetsKVStore is transient.
// An example is creating a new NetsKVStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// NetsKVStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (bs *NatsStore) StopCleanup() {
	if bs.stopCleanup != nil {
		bs.stopCleanup <- true
	}
}

// All implements scs.IterableStore.
func (ns *NatsStore) All() (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()
	return ns.AllCtx(ctx)
}

// FindCtx implements scs.CtxStore.
func (ns *NatsStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	return ns.get(ctx, token, time.Now())
}

// Find implements scs.Store.
func (ns *NatsStore) Find(token string) ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()
	return ns.FindCtx(ctx, token)
}

// CommitCtx implements scs.CtxStore.
func (ns *NatsStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	return ns.put(ctx, token, b, expiry)
}

// Commit implements scs.Store.
func (ns *NatsStore) Commit(token string, b []byte, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()
	return ns.CommitCtx(ctx, token, b, expiry)
}

// DeleteCtx implements scs.CtxStore.
func (ns *NatsStore) DeleteCtx(ctx context.Context, token string) error {
	return ns.delete(ctx, token)
}

func (ns *NatsStore) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()
	return ns.DeleteCtx(ctx, token)
}

// New creates a NatsStore instance. db should be a pointer to a jetstreamKeyValue store.
func New(db jetstream.KeyValue, opts ...Opt) *NatsStore {
	ns := &NatsStore{db: db, cleanup: time.Minute}
	for _, opt := range opts {
		opt(ns)
	}
	if ns.cleanup > 0 {
		go ns.StartCleanup()
	}
	return ns
}

type Opt func(ns *NatsStore)

func WithTimeout(t time.Duration) Opt {
	return func(ns *NatsStore) {
		ns.timeout = t
	}
}

// CleanupFrequency sets how frequently stale session data gets cleaned up. It's 1 min by default
func WithCleanupInterval(t time.Duration) Opt {
	return func(ns *NatsStore) {
		ns.cleanup = t
	}
}
