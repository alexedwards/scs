package sqlite3store

import (
	"database/sql"
	"log"
	"time"
)

// SQLite3Store represents the session store.
type SQLite3Store struct {
	db          *sql.DB
	stopCleanup chan bool
}

// New returns a new SQLite3Store instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *sql.DB) *SQLite3Store {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

// NewWithCleanupInterval returns a new SQLite3Store instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) *SQLite3Store {
	p := &SQLite3Store{db: db}
	if cleanupInterval > 0 {
		go p.startCleanup(cleanupInterval)
	}
	return p
}

// Find returns the data for a given session token from the SQLite3Store instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *SQLite3Store) Find(token string) (b []byte, exists bool, err error) {
	row := p.db.QueryRow("SELECT data FROM sessions WHERE token = $1 AND julianday('now') < expiry", token)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the SQLite3Store instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *SQLite3Store) Commit(token string, b []byte, expiry time.Time) error {
	_, err := p.db.Exec("REPLACE INTO sessions (token, data, expiry) VALUES ($1, $2, julianday($3))", token, b, expiry)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the SQLite3Store
// instance.
func (p *SQLite3Store) Delete(token string) error {
	_, err := p.db.Exec("DELETE FROM sessions WHERE token = $1", token)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the SQLite3Store instance.
func (p *SQLite3Store) All() (map[string][]byte, error) {
	rows, err := p.db.Query("SELECT token, data FROM sessions WHERE julianday('now') < expiry")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make(map[string][]byte)

	for rows.Next() {
		var (
			token string
			data  []byte
		)

		err = rows.Scan(&token, &data)
		if err != nil {
			return nil, err
		}

		sessions[token] = data
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (p *SQLite3Store) startCleanup(interval time.Duration) {
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

// StopCleanup terminates the background cleanup goroutine for the SQLite3Store
// instance. It's rare to terminate this; generally SQLite3Store instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the SQLite3Store is transient.
// An example is creating a new SQLite3Store instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// SQLite3Store object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (p *SQLite3Store) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *SQLite3Store) deleteExpired() error {
	_, err := p.db.Exec("DELETE FROM sessions WHERE expiry < julianday('now')")
	return err
}
