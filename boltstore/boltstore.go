package boltstore

import (
	"encoding/binary"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

var bucketName = []byte("scs:session")

// BoltStore represents the session store.
type BoltStore struct {
	db          *bbolt.DB
	stopCleanup chan bool
}

// New returns a new Boltstore instance, with a background cleanup goroutine
// that runs every 1 minute to remove expired session data.
func New(db *bbolt.DB) *BoltStore {
	return NewWithCleanupInterval(db, time.Minute)
}

// NewWithCleanupInterval returns a new Boltstore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *bbolt.DB, cleanupInterval time.Duration) *BoltStore {
	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	bs := &BoltStore{
		db: db,
	}
	if cleanupInterval > 0 {
		go bs.startCleanup(cleanupInterval)
	}
	return bs
}

// Find returns the data for a given session token from the BoltStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (bs *BoltStore) Find(token string) (b []byte, exists bool, err error) {
	var val []byte
	err = bs.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		val = bucket.Get([]byte(token))
		if val == nil {
			return nil
		}

		if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(val[:8]) {
			val = nil
		}

		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if val == nil {
		return nil, false, nil
	}
	return val[8:], true, err
}

// Commit adds a session token and data to the BoltStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (bs *BoltStore) Commit(token string, b []byte, expiry time.Time) error {
	return bs.db.Update(func(tx *bbolt.Tx) error {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(expiry.UnixNano()))
		val := append(buf, b...)

		bucket := tx.Bucket(bucketName)
		err := bucket.Put([]byte(token), val)
		return err
	})
}

// Delete removes a session token and corresponding data from the BoltStore
// instance.
func (bs *BoltStore) Delete(token string) error {
	return bs.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		return bucket.Delete([]byte(token))
	})
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the BoltStore instance.
func (bs *BoltStore) All() (map[string][]byte, error) {
	sessions := make(map[string][]byte)

	err := bs.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		cursor := bucket.Cursor()

		for key, val := cursor.First(); key != nil; key, val = cursor.Next() {
			if binary.BigEndian.Uint64(val[:8]) > uint64(time.Now().UnixNano()) {
				sessions[string(key)] = val[8:]
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (bs *BoltStore) startCleanup(cleanupInterval time.Duration) {
	bs.stopCleanup = make(chan bool)
	ticker := time.NewTicker(cleanupInterval)
	for {
		select {
		case <-ticker.C:
			err := bs.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-bs.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the BoltStore
// instance. It's rare to terminate this; generally BoltStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the BoltStore is transient.
// An example is creating a new BoltStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// BoltStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (bs *BoltStore) StopCleanup() {
	if bs.stopCleanup != nil {
		bs.stopCleanup <- true
	}
}

func (bs *BoltStore) deleteExpired() error {
	var expiredTokens [][]byte
	bs.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		bucket.ForEach(func(token, val []byte) error {
			if uint64(time.Now().UnixNano()) > binary.BigEndian.Uint64(val[:8]) {
				expiredTokens = append(expiredTokens, token)
			}
			return nil
		})
		return nil
	})

	if len(expiredTokens) > 0 {
		return bs.db.Update(func(tx *bbolt.Tx) error {
			for _, token := range expiredTokens {
				bucket := tx.Bucket(bucketName)
				return bucket.Delete([]byte(token))
			}
			return nil
		})
	}

	return nil
}
