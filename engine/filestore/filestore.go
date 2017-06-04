// Package filestore is a file-based storage engine for the SCS session package.
//
// Warning: Because filestore uses file based storage it is slow and should not
// be used in production.  It mearly exists to provide the convenience of
// memstore while maintaining data across server restarts.
//
// **dev only don't use in production**

// The filestore package provides a background 'cleanup' goroutine to delete
// expired session data. This stops the underlying cache from holding on to invalid
// sessions forever and taking up unnecessary memory.
//
// Usage:
//
//	func main() {
//		// Create a new filestore storage engine with a cleanup interval of 5 minutes.
//		engine := filestore.New("/tmp/cookies.data", 5 * time.Minute)
//
//		sessionManager := session.Manage(engine)
//		http.ListenAndServe(":4000", sessionManager(http.DefaultServeMux))
//	}
package filestore

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var errTypeAssertionFailed = errors.New("type assertion failed: could not convert interface{} to []byte")

// FileStore represents the currently configured session storage engine. It is essentially
// a wrapper around a go-cache instance (see https://github.com/patrickmn/go-cache).
type FileStore struct {
	*cache.Cache
	filePath string
}

// New returns a new FileStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background 'cleanup' goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(filePath string, cleanupInterval time.Duration) *FileStore {
	m := &FileStore{
		Cache:    cache.New(cache.DefaultExpiration, cleanupInterval),
		filePath: filePath,
	}
	m.LoadFile(m.filePath)
	return m
}

// Find returns the data for a given session token from the FileStore instance. If the session
// token is not found or is expired, the returned exists flag will be set to false.
func (m *FileStore) Find(token string) (b []byte, exists bool, err error) {
	v, exists := m.Cache.Get(token)
	if exists == false {
		return nil, exists, nil
	}

	b, ok := v.([]byte)
	if ok == false {
		return nil, exists, errTypeAssertionFailed
	}

	return b, exists, nil
}

// Save adds a session token and data to the FileStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (m *FileStore) Save(token string, b []byte, expiry time.Time) error {
	m.Cache.Set(token, b, expiry.Sub(time.Now()))
	return m.SaveFile(m.filePath)
}

// Delete removes a session token and corresponding data from the FileStore instance.
func (m *FileStore) Delete(token string) error {
	m.Cache.Delete(token)
	return m.SaveFile(m.filePath)
}
