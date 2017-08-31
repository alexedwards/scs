// Package buntstore is a buntdb based session store for the SCS session package.
package buntstore

import (
	"time"

	"github.com/tidwall/buntdb"
)

// BuntStore is a SCS session store backed by a buntdb file.
type BuntStore struct {
	db *buntdb.DB
}

// New creates a BuntStore instance.
func New(db *buntdb.DB) *BuntStore {
	store := &BuntStore{
		db: db,
	}
	return store
}

// Save updates data for a given session token with a given expiry.
// Any existing data + expiry will be over-written.
func (bs *BuntStore) Save(token string, b []byte, expiry time.Time) error {
	return bs.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(token, string(b), &buntdb.SetOptions{Expires: true, TTL: expiry.Sub(time.Now())})
		return err
	})
}

// Find returns the data for a session token.
// If the session token is not found or is expired,
// the exists flag will be false.
func (bs *BuntStore) Find(token string) (b []byte, exists bool, err error) {
	var value string
	err = bs.db.View(func(tx *buntdb.Tx) error {
		value, err = tx.Get(token)
		return err
	})
	if err != nil {
		if err == buntdb.ErrNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return []byte(value), value != "", err
}

// Delete removes session token and corresponding data.
func (bs *BuntStore) Delete(token string) error {
	return bs.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(token)
		return err
	})
}
