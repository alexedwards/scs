package scs

import (
	"context"
	"time"
)

// Store is the interface for session stores.
type Store interface {
	// Delete should remove the session token and corresponding data from the
	// session store. If the token does not exist then Delete should be a no-op
	// and return nil (not an error).
	Delete(token string) (err error)

	// Find should return the data for a session token from the store. If the
	// session token is not found or is expired, the found return value should
	// be false (and the err return value should be nil). Similarly, tampered
	// or malformed tokens should result in a found return value of false and a
	// nil err value. The err return value should be used for system errors only.
	Find(token string) (b []byte, found bool, err error)

	// Commit should add the session token and data to the store, with the given
	// expiry time. If the session token already exists, then the data and
	// expiry time should be overwritten.
	Commit(token string, b []byte, expiry time.Time) (err error)
}

// IterableStore is the interface for session stores which support iteration.
type IterableStore interface {
	// All should return a map containing data for all active sessions (i.e.
	// sessions which have not expired). The map key should be the session
	// token and the map value should be the session data. If no active
	// sessions exist this should return an empty (not nil) map.
	All() (map[string][]byte, error)
}

// CtxStore is an interface for session stores which take a context.Context
// parameter.
type CtxStore interface {
	Store

	// DeleteCtx is the same as Store.Delete, except it takes a context.Context.
	DeleteCtx(ctx context.Context, token string) (err error)

	// FindCtx is the same as Store.Find, except it takes a context.Context.
	FindCtx(ctx context.Context, token string) (b []byte, found bool, err error)

	// CommitCtx is the same as Store.Commit, except it takes a context.Context.
	CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error)
}

// IterableCtxStore is the interface for session stores which support iteration
// and which take a context.Context parameter.
type IterableCtxStore interface {
	// AllCtx is the same as IterableStore.All, expect it takes a
	// context.Context.
	AllCtx(ctx context.Context) (map[string][]byte, error)
}
