// Package pgstore is a PostgreSQL-based session store for the SCS session package.
//
// A working PostgreSQL database is required, containing a sessions table with
// the definition:
//
//	CREATE TABLE sessions (
//	  token TEXT PRIMARY KEY,
//	  data BYTEA NOT NULL,
//	  expiry TIMESTAMPTZ NOT NULL
//	);
//	CREATE INDEX sessions_expiry_idx ON sessions (expiry);
//
// The pgstore package provides a background 'cleanup' goroutine to delete expired
// session data. This stops the database table from holding on to invalid sessions
// indefinitely and growing unnecessarily large.
package pgstore

import (
	"database/sql"
	"log"
	"time"

	// Register lib/pq with database/sql
	_ "github.com/lib/pq"
)

// PGStore represents the currently configured session session store.
type PGStore struct {
	*sql.DB
	stopCleanup chan bool
}

// New returns a new PGStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background cleanup goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(db *sql.DB, cleanupInterval time.Duration) *PGStore {
	p := &PGStore{DB: db}
	if cleanupInterval > 0 {
		go p.startCleanup(cleanupInterval)
	}
	return p
}

// Find returns the data for a given session token from the PGStore instance. If
// the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *PGStore) Find(token string) (b []byte, exists bool, err error) {
	row := p.DB.QueryRow("SELECT data FROM sessions WHERE token = $1 AND current_timestamp < expiry", token)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Save adds a session token and data to the PGStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (p *PGStore) Save(token string, b []byte, expiry time.Time) error {
	_, err := p.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, $3) ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry", token, b, expiry)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the PGStore instance.
func (p *PGStore) Delete(token string) error {
	_, err := p.DB.Exec("DELETE FROM sessions WHERE token = $1", token)
	return err
}

func (p *PGStore) startCleanup(interval time.Duration) {
	p.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := p.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-p.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the PGStore instance.
// It's rare to terminate this; generally PGStore instances and their cleanup
// goroutines are intended to be long-lived and run for the lifetime of  your
// application.
//
// There may be occasions though when your use of the PGStore is transient. An
// example is creating a new PGStore instance in a test function. In this scenario,
// the cleanup goroutine (which will run forever) will prevent the PGStore object
// from being garbage collected even after the test function has finished. You
// can prevent this by manually calling StopCleanup.
func (p *PGStore) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *PGStore) deleteExpired() error {
	_, err := p.DB.Exec("DELETE FROM sessions WHERE expiry < current_timestamp")
	return err
}
