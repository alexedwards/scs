package rqlitestore

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/rqlite/gorqlite"
)

// RqliteStore represents the session store.
type RqliteStore struct {
	conn        *gorqlite.Connection
	stopCleanup chan bool
}

// New returns a new RqliteStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(conn gorqlite.Connection) *RqliteStore {
	return NewWithCleanupInterval(conn, 5*time.Minute)
}

// NewWithCleanupInterval returns a new RqliteStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(conn gorqlite.Connection, cleanupInterval time.Duration) *RqliteStore {
	r := &RqliteStore{conn: &conn}

	if cleanupInterval > 0 {
		go r.startCleanup(cleanupInterval)
	}

	return r
}

// Find returns the data for a given session token from the RqliteStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (r *RqliteStore) Find(token string) (b []byte, exists bool, err error) {
	query := fmt.Sprintf("SELECT data FROM sessions WHERE token = '%s' AND julianday('now') < expiry", token)
	row, err := r.conn.QueryOne(query)
	if err != nil {
		return nil, false, err
	}

	for row.Next() {
		var datax string

		err := row.Scan(&datax)
		if err != nil {
			return nil, false, err
		}

		b, err = hex.DecodeString(datax)
		if err != nil {
			return nil, false, err
		}
	}
	if row.NumRows() == 0 {
		return nil, false, nil
	} else if row.Err != nil {
		return nil, false, row.Err
	}

	return b, true, nil
}

// Commit adds a session token and data to the RqliteStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (r *RqliteStore) Commit(token string, b []byte, expiry time.Time) error {
	query := fmt.Sprintf("REPLACE INTO sessions (token, data, expiry) VALUES ('%s', '%x', julianday('%s'))", token, b, expiry.UTC().Format("2006-01-02T15:04:05.000"))
	_, err := r.conn.WriteOne(query)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a session token and corresponding data from the RqliteStore
// instance.
func (r *RqliteStore) Delete(token string) error {
	query := fmt.Sprintf("DELETE FROM sessions WHERE token = '%s'", token)
	_, err := r.conn.WriteOne(query)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the RqliteStore instance.
func (r *RqliteStore) All() (map[string][]byte, error) {
	rows, err := r.conn.QueryOne("SELECT token, data FROM sessions WHERE julianday('now') < expiry")
	if err != nil {
		return nil, err
	}

	sessions := make(map[string][]byte)

	for rows.Next() {
		var (
			token string
			datax string
			data  []byte
		)

		err = rows.Scan(&token, &datax)
		if err != nil {
			return nil, err
		}

		data, err = hex.DecodeString(datax)
		if err != nil {
			return nil, err
		}

		sessions[token] = data
	}
	if rows.Err != nil {
		return nil, rows.Err
	}

	return sessions, nil
}

func (r *RqliteStore) startCleanup(interval time.Duration) {
	r.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			err := r.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-r.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the RqliteStore
// instance. It's rare to terminate this; generally RqliteStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the RqliteStore is transient.
// An example is creating a new RqliteStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// RqliteStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (r *RqliteStore) StopCleanup() {
	if r.stopCleanup != nil {
		r.stopCleanup <- true
	}
}

func (r *RqliteStore) deleteExpired() error {
	_, err := r.conn.WriteOne("DELETE FROM sessions WHERE expiry < julianday('now')")
	return err
}
