package buntdbstore

import (
	"time"

	"github.com/tidwall/buntdb"
)

// BuntDBStore represents the session store.
type BuntDBStore struct {
	db *buntdb.DB
}

// New returns a new BuntDBStore instance.
// The db parameter should be a pointer to a buntdb store instance.
func New(db *buntdb.DB) *BuntDBStore {
	store := &BuntDBStore{
		db: db,
	}
	return store
}

// Find returns the data for a given session token from the BuntDBStore
// instance. If the session token is not found or is expired,
// the returned exists flag will be set to false.
func (bs *BuntDBStore) Find(token string) (b []byte, exists bool, err error) {
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

// Commit adds a session token and data to the BuntDBStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (bs *BuntDBStore) Commit(token string, b []byte, expiry time.Time) error {
	return bs.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(token, string(b), &buntdb.SetOptions{Expires: true, TTL: expiry.Sub(time.Now())})
		return err
	})
}

// Delete removes a session token and corresponding data from the BuntDBStore instance.
func (bs *BuntDBStore) Delete(token string) error {
	return bs.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(token)
		return err
	})
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the BuntDBStore instance.
func (bs *BuntDBStore) All() (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	err := bs.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			sessions[key] = []byte(value)
			return true
		})
		return err
	})
	if err != nil {
		if err == buntdb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return sessions, nil
}
