package badgerstore

import (
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerStore represents the session store.
type BadgerStore struct {
	db     *badger.DB
	prefix string
}

// New returns a new BadgerStore instance.
// The db parameter should be a pointer to a badger store instance.
func New(db *badger.DB) *BadgerStore {
	return NewWithPrefix(db, "scs:session:")
}

// NewWithPrefix returns a new BadgerStore instance.
// The db parameter should be a pointer to a badger store instance.
// The prefix parameter controls the Badger key prefix,
// which can be used to avoid naming clashes if necessary.
func NewWithPrefix(db *badger.DB, prefix string) *BadgerStore {
	return &BadgerStore{
		db:     db,
		prefix: prefix,
	}
}

// Find returns the data for a given session token from the BadgerStore
// instance. If the session token is not found or is expired,
// the returned exists flag will be set to false.
func (bs *BadgerStore) Find(token string) ([]byte, bool, error) {
	key := []byte(bs.prefix + token)
	txn := bs.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	data, err := item.ValueCopy(nil)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

// Commit adds a session token and data to the BadgerStore instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (bs *BadgerStore) Commit(token string, data []byte, expiry time.Time) error {
	txn := bs.db.NewTransaction(true)
	defer txn.Discard()

	key := []byte(bs.prefix + token)
	entry := badger.NewEntry(key, data).WithTTL(time.Until(expiry))
	err := txn.SetEntry(entry)
	if err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

// Delete removes a session token and corresponding data from the BadgerStore instance.
func (bs *BadgerStore) Delete(token string) error {
	txn := bs.db.NewTransaction(true)
	defer txn.Discard()

	txn.Delete([]byte(bs.prefix + token))

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the BadgerStore instance.
func (bs *BadgerStore) All() (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	err := bs.db.View(func(txn *badger.Txn) error {
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()

		prefix := []byte(bs.prefix)
		for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
			item := iterator.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				token := string(key)[len(prefix):]
				sessions[token] = val
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
