// Package xormstore Provide xorm store by xorm
package xormstore

import (
	"log"
	"time"

	"xorm.io/xorm"
)

type session struct {
	Expiry time.Time `xorm:"'expiry' index(sessions_expiry_idx) not null"`
	Token  string    `xorm:"'token' pk varchar(43)"`
	Data   []byte    `xorm:"'data' not null"`
}

func (session) TableName() string {
	return "sessions"
}

// XormStore represents the session store.
type XormStore struct {
	db          *xorm.Engine
	stopCleanup chan bool
}

// New return a new XormStore instance with a background cleanup
// goroutine that runs every 5 minutes to remove expired session data
func New(db *xorm.Engine) (*XormStore, error) {
	return NewWithCleanupInterval(db, time.Minute*5)
}

// NewWithCleanupInterval returns a new XormStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *xorm.Engine, cleanupInterval time.Duration) (*XormStore, error) {
	m := &XormStore{
		db: db,
	}

	if err := m.db.Sync(&session{}); err != nil {
		return nil, err
	}

	if cleanupInterval > 0 {
		go m.startCleanup(cleanupInterval)
	}

	return m, nil
}

// Find returns the data for a given session token from the XormStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (xs *XormStore) Find(token string) (b []byte, exists bool, err error) {
	s := &session{}
	has, err := xs.db.Where("token = ?", token).And("expiry >= ?", time.Now()).Get(s)
	if err != nil {
		return nil, false, err
	}

	if !has {
		return nil, false, nil
	}

	return s.Data, true, nil
}

// Commit adds a session token and data to the GORMStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (xs *XormStore) Commit(token string, b []byte, expiry time.Time) error {
	s := &session{Token: token}

	has, err := xs.db.Get(s)
	if err != nil {
		return err
	}
	if !has {
		s.Data = b
		s.Expiry = expiry

		_, err := xs.db.Insert(s)
		if err != nil {
			return err
		}

		return nil
	}

	_, err = xs.db.Where("token = ?", token).Update(&session{Data: b, Expiry: expiry})
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a session token and corresponding data from the XormStore
// instance.
func (xs *XormStore) Delete(token string) error {
	_, err := xs.db.Where("token = ?", token).Delete(&session{})

	return err
}

// All return all not expired session
func (xs *XormStore) All() (map[string][]byte, error) {
	rows, err := xs.db.Where("expiry >= ?", time.Now()).Rows(&session{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ss := make(map[string][]byte)
	for rows.Next() {
		s := &session{}
		if err := rows.Scan(s); err != nil {
			return nil, err
		}

		ss[s.Token] = s.Data
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ss, nil
}

// startCleanup start cleanup goroutine
func (xs *XormStore) startCleanup(interval time.Duration) {
	xs.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			err := xs.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-xs.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup stop cleanup goroutine
func (xs *XormStore) StopCleanup() {
	if xs.stopCleanup != nil {
		xs.stopCleanup <- true
	}
}

func (xs *XormStore) deleteExpired() error {
	_, err := xs.db.Where("expiry < ?", time.Now()).Delete(&session{})

	return err
}
