package etcdstore

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdStore represents the session store.
type EtcdStore struct {
	client *clientv3.Client
	prefix string
}

// New returns a new EtcdStore instance.
// The client parameter should be a pointer to a etcd client instance.
func New(client *clientv3.Client) *EtcdStore {
	return NewWithPrefix(client, "scs:session:")
}

// NewWithPrefix returns a new EtcdStore instance. The client parameter should be a pointer
// to a etcd client instance. The prefix parameter controls the etcd key
// prefix, which can be used to avoid naming clashes if necessary.
func NewWithPrefix(client *clientv3.Client, prefix string) *EtcdStore {
	return &EtcdStore{
		client: client,
		prefix: prefix,
	}
}

// FindCtx returns the data for a given session token from the EtcdStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (e *EtcdStore) FindCtx(ctx context.Context, token string) (b []byte, exists bool, err error) {
	res, err := e.client.Get(ctx, e.prefix+token)
	if err != nil {
		return nil, false, err
	}

	if len(res.Kvs) == 0 {
		return nil, false, nil
	}

	return res.Kvs[0].Value, true, nil
}

// CommitCtx adds a session token and data to the EtcdStore instance with the
// given expiry time. If the session token already exists then the data and expiry
// time are updated.
func (e *EtcdStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	lease, _ := e.client.Grant(ctx, int64(time.Until(expiry).Seconds()))
	_, err := e.client.Put(ctx, e.prefix+token, string(b), clientv3.WithLease(lease.ID))
	return err
}

// DeleteCtx removes a session token and corresponding data from the EtcdStore
// instance.
func (e *EtcdStore) DeleteCtx(ctx context.Context, token string) error {
	_, err := e.client.Delete(ctx, e.prefix+token)
	return err
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the EtcdStore instance.
func (e *EtcdStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}

	res, err := e.client.Get(ctx, e.prefix, opts...)
	if err != nil {
		return nil, err
	}

	if len(res.Kvs) == 0 {
		return sessions, nil
	}

	for _, kv := range res.Kvs {
		sessions[string(kv.Key)[len(e.prefix):]] = kv.Value
	}

	return sessions, nil
}

// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a seperate type makes any errors caught
// only at runtime instead of compile time. Oh well.

func (e *EtcdStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}
func (e *EtcdStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}
func (e *EtcdStore) Delete(token string) error {
	panic("missing context arg")
}
