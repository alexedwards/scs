// Package boltstore is a boltdb based session store for the SCS session package.
package boltstore

import (
	"log"
	"time"

	"github.com/boltdb/bolt"
)

var (
	dataBucketName   = []byte("scs_data_bucket")
	expiryBucketName = []byte("scs_expiry_bucket")
)

// BoltStore is a SCS session store backed by a boltdb file.
type BoltStore struct {
	db          *bolt.DB
	stopCleanup chan bool
}

// New creates a BoltStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background cleanup goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(db *bolt.DB, cleanupInterval time.Duration) *BoltStore {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(dataBucketName)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(expiryBucketName)
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

// Save updates data for a given session token with a given expiry.
// Any existing data + expiry will be over-written.
func (bs *BoltStore) Save(token string, b []byte, expiry time.Time) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		tokenBytes := []byte(token)

		bucket := tx.Bucket(dataBucketName)
		err := bucket.Put(tokenBytes, b)
		if err != nil {
			return err
		}

		expiryBucket := tx.Bucket(expiryBucketName)
		expBytes, err := expiry.MarshalText()
		if err != nil {
			return err
		}
		return expiryBucket.Put(tokenBytes, expBytes)
	})
}

// Find returns the data for a session token.
// If the session token is not found or is expired,
// the exists flag will be false.
func (bs *BoltStore) Find(token string) (b []byte, exists bool, err error) {
	var value []byte
	err = bs.db.View(func(tx *bolt.Tx) error {
		tokenBytes := []byte(token)

		bucket := tx.Bucket(dataBucketName)
		value = bucket.Get(tokenBytes)

		if value == nil {
			return nil
		}

		expiryBucket := tx.Bucket(expiryBucketName)
		expiryBytes := expiryBucket.Get(tokenBytes)

		if isExpired(expiryBytes) {
			value = nil
		}

		return nil
	})
	return value, value != nil, err
}

// Delete removes session token and corresponding data.
func (bs *BoltStore) Delete(token string) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		tokenBytes := []byte(token)
		return txDelete(tx, tokenBytes)
	})
}

// startCleanup is a helper func to periodically call deleteExpired.
// It will stop if/when it recieves a message on stopCleanup channel.
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

// StopCleanup terminates the background cleanup goroutine for the BoltStore instance.
// It's rare to terminate this; generally BoltStore instances and their cleanup
// goroutines are intended to be long-lived and run for the lifetime of  your
// application.
//
// There may be occasions though when your use of the BoltStore is transient. An
// example is creating a new BoltStore instance in a test function. In this scenario,
// the cleanup goroutine (which will run forever) will prevent the BoltStore object
// from being garbage collected even after the test function has finished. You
// can prevent this by manually calling StopCleanup.
func (bs *BoltStore) StopCleanup() {
	if bs.stopCleanup != nil {
		bs.stopCleanup <- true
	}
}

// deleteExpired runs at in a separate goroutine at cleanupInterval
// as specified in the New constructor.
//
// iterate over keys in the expiry bucket,
// and delete keys that are exipred.
func (bs *BoltStore) deleteExpired() error {
	var expiredKeys [][]byte
	bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(expiryBucketName)
		b.ForEach(func(k, v []byte) error {
			if isExpired(v) {
				expiredKeys = append(expiredKeys, k)
			}
			return nil
		})
		return nil
	})

	if len(expiredKeys) > 0 {
		return bs.db.Update(func(tx *bolt.Tx) error {
			for _, k := range expiredKeys {
				if err := txDelete(tx, k); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return nil
}

// txDelete is a helper to delete a key
// from both the data + expiry bucket
// inside a transaction.
func txDelete(tx *bolt.Tx, tokenBytes []byte) error {
	expiryBucket := tx.Bucket(expiryBucketName)
	expiryBucket.Delete(tokenBytes)

	bucket := tx.Bucket(dataBucketName)
	return bucket.Delete(tokenBytes)
}

// isExpired is a helper func to unmarshal a expiry date
// and determine if it is after Now.
func isExpired(expiryBytes []byte) bool {
	expiry := &time.Time{}
	err := expiry.UnmarshalText(expiryBytes)
	if err != nil {
		return true
	}
	return time.Now().After(*expiry)
}
