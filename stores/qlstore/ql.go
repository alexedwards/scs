// Package qlstore is a ql-based session store for the SCS session package.
//
// A working ql database is required, containing a sessions table with
// the definition:
//
//	CREATE TABLE sessions (
//		token string,
//		data blob,
//		expiry time
//	)
//	CREATE INDEX sessions_expiry_idx ON sessions (expiry);
//
// The qlstore package provides a background 'cleanup' goroutine to delete expired
// session data. This stops the database table from holding on to invalid sessions
// indefinitely and growing unnecessarily large.
package qlstore

import (
	"database/sql"
	"log"
	"time"

	// Register ql driver with database/sql
	_ "github.com/cznic/ql/driver"
)

// QLStore represents the currently configured session session store.
type QLStore struct {
	*sql.DB
	stopCleanup chan bool
}

// New returns a new QLStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background cleanup goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(db *sql.DB, cleanupInterval time.Duration) *QLStore {
	q := &QLStore{DB: db}
	if cleanupInterval > 0 {
		go q.startCleanup(cleanupInterval)
	}
	return q
}

func (q *QLStore) startCleanup(interval time.Duration) {
	q.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := q.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-q.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// Delete removes a session token and corresponding data from the QLStore instance.
func (q *QLStore) Delete(token string) error {
	_, err := execTx(q.DB, "DELETE FROM sessions where token=$1", token)
	return err
}

func (q *QLStore) deleteExpired() error {
	_, err := execTx(q.DB, "DELETE FROM sessions WHERE expiry < now()")
	return err
}

// Find returns the data for a given session token from the QLStore instance. If
// the session token is not found or is expired, the returned exists flag will
// be set to false.
func (q *QLStore) Find(token string) ([]byte, bool, error) {
	var data []byte
	query := "SELECT data FROM sessions WHERE token=$1 AND now()<expiry"
	err := q.QueryRow(query, token).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

// Save adds a session token and data to the QLStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (q *QLStore) Save(token string, b []byte, expiry time.Time) error {
	_, ok, _ := q.Find(token)
	if ok {
		_, err := execTx(q.DB, `
		UPDATE sessions data=$2,expiry=$3 WHERE token=$1
		`, token, b, expiry)
		return err
	}
	_, err := execTx(q.DB, `
	INSERT INTO sessions (token , data, expiry) VALUES ($1,$2,$3)
	`, token, b, expiry)
	return err
}

func execTx(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Commit()
	}()
	r, err := tx.Exec(query, args...)
	return r, err
}

// StopCleanup terminates the background cleanup goroutine for the QLStore instance.
// It's rare to terminate this; generally QLStore instances and their cleanup
// goroutines are intended to be long-lived and run for the lifetime of  your
// application.
//
// There may be occasions though when your use of the QLStore is transient. An
// example is creating a new QLStore instance in a test function. In this scenario,
// the cleanup goroutine (which will run forever) will prevent the QLStore object
// from being garbage collected even after the test function has finished. You
// can prevent this by manually calling StopCleanup.
func (q *QLStore) StopCleanup() {
	if q.stopCleanup != nil {
		q.stopCleanup <- true
	}
}

//Table provides SQL for creating a session table in ql database
func Table() string {
	return `
	CREATE TABLE sessions (
		token string,
		data blob,
		expiry time
	)
	`
}
