package coherencestore

import (
	"context"
	"errors"
	"fmt"
	"github.com/oracle/coherence-go-client/coherence"
	"time"
)

const defaultCacheName = "default-session-store"

// CoherenceStore represents a Coherence store used.
type CoherenceStore struct {
	session    *coherence.Session
	namedCache coherence.NamedCache[string, []byte]
}

// New Returns a new Session Store using the default cache name.
func New(session *coherence.Session) (*CoherenceStore, error) {
	return newCoherenceStore(session, defaultCacheName)
}

// NewWithCache returns a new Session Store using a custom cache name.
func NewWithCache(session *coherence.Session, cacheName string) (*CoherenceStore, error) {
	return newCoherenceStore(session, cacheName)
}

func newCoherenceStore(session *coherence.Session, cacheName string) (*CoherenceStore, error) {
	nc, err := coherence.GetNamedCache[string, []byte](session, "scs$"+cacheName)
	if err != nil {
		return nil, err
	}
	return &CoherenceStore{
		session:    session,
		namedCache: nc,
	}, nil
}

func (c *CoherenceStore) String() string {
	return fmt.Sprintf("CoherenceStore{cacheName=%v,session=%v}", c.namedCache.Name(), c.session)
}

// CommitCtx saves the data for a session token with a timeout using a specific [context.Context].
func (c *CoherenceStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error) {
	// figure out the ttl based upon the expiry time
	ttl := expiry.Sub(time.Now()).Milliseconds()
	if ttl < 0 {
		return errors.New("time is before now")
	}
	_, err = c.namedCache.PutWithExpiry(ctx, token, b, time.Duration(ttl)*time.Millisecond)
	return err
}

// FindCtx returns the data for a session token using a specific [context.Context].
func (c *CoherenceStore) FindCtx(ctx context.Context, token string) (b []byte, found bool, err error) {
	v, err := c.namedCache.Get(ctx, token)
	if err != nil {
		return nil, false, err
	}
	if v == nil {
		return nil, false, err
	}
	return *v, true, nil
}

// DeleteCtx removes the session token using a specific [context.Context].
func (c *CoherenceStore) DeleteCtx(ctx context.Context, token string) (err error) {
	_, err = c.namedCache.Remove(ctx, token)
	return err
}

// AllCtx returns a map containing data for all active sessions using a specific [context.Context]
// (i.e. sessions which have not expired).
func (c *CoherenceStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	sessions := make(map[string][]byte, 0)
	for e := range c.namedCache.EntrySet(ctx) {
		if e.Err != nil {
			return sessions, e.Err
		}
		sessions[e.Key] = e.Value
	}

	return sessions, nil
}

// Commit saves the data for a session token with a timeout.
func (c *CoherenceStore) Commit(token string, b []byte, expiry time.Time) error {
	return c.CommitCtx(context.Background(), token, b, expiry)
}

// Find returns the data for a session token.
func (c *CoherenceStore) Find(token string) ([]byte, bool, error) {
	return c.FindCtx(context.Background(), token)
}

// Delete removes the session token.
func (c *CoherenceStore) Delete(token string) error {
	return c.DeleteCtx(context.Background(), token)
}

// All returns a map containing data for all active sessions (i.e.
// sessions which have not expired).
func (c *CoherenceStore) All() (map[string][]byte, error) {
	return c.AllCtx(context.Background())
}
