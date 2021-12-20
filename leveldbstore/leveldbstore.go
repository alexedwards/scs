package leveldbstore

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var basePrefix string = "scs:session:"

// LevelDBStore represents the session store.
type LevelDBStore struct {
	db          *leveldb.DB
	stopCleanup chan bool
}

// New returns a new LevelDBStore instance, with a background cleanup goroutine
// that runs every 1 minute to remove expired session data.
func New(db *leveldb.DB) *LevelDBStore {
	return NewWithCleanupInterval(db, time.Minute)
}

// NewWithCleanupInterval returns a new LevelDBStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *leveldb.DB, cleanupInterval time.Duration) *LevelDBStore {
	bs := &LevelDBStore{
		db: db,
	}

	if cleanupInterval > 0 {
		go bs.startCleanup(cleanupInterval)
	}

	return bs
}

// Find returns the data for a given session token from the LevelDBStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (ls *LevelDBStore) Find(token string) (b []byte, exists bool, err error) {
	var val []byte

	val, err = ls.db.Get([]byte(basePrefix+token), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(val[:8]) {
		val = nil
	}
	if val == nil {
		return nil, false, nil
	}

	return val[8:], true, nil
}

// Commit adds a session token and data to the LevelDBStore instance with the
// given expiry time. If the session token already exists then the data and expiry
// time are updated.
func (ls *LevelDBStore) Commit(token string, b []byte, expiry time.Time) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(expiry.UnixNano()))
	val := append(buf, b...)

	return ls.db.Put([]byte(basePrefix+token), val, nil)
}

// Delete removes a session token and corresponding data from the LevelDBStore
// instance.
func (ls *LevelDBStore) Delete(token string) error {
	return ls.db.Delete([]byte(basePrefix+token), nil)
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the LevelDBStore instance.
func (ls *LevelDBStore) All() (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	iter := ls.db.NewIterator(util.BytesPrefix([]byte(basePrefix)), nil)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		if binary.BigEndian.Uint64(val[:8]) > uint64(time.Now().UnixNano()) {
			sessions[string(key[len(basePrefix):])] = val[8:]
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (ls *LevelDBStore) startCleanup(cleanupInterval time.Duration) {
	ls.stopCleanup = make(chan bool)
	ticker := time.NewTicker(cleanupInterval)
	for {
		select {
		case <-ticker.C:
			err := ls.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-ls.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the LevelDBStore
// instance. It's rare to terminate this; generally LevelDBStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the LevelDBStore is transient.
// An example is creating a new LevelDBStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// LevelDBStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (ls *LevelDBStore) StopCleanup() {
	if ls.stopCleanup != nil {
		ls.stopCleanup <- true
	}
}

func (ls *LevelDBStore) deleteExpired() error {
	iter := ls.db.NewIterator(util.BytesPrefix([]byte(basePrefix)), nil)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(val[:8]) {
			return ls.db.Delete([]byte(key), nil)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}
