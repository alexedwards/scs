package mssqlstore

import (
	"database/sql"
	"log"
	"time"
)

// MSSQLStore represents the session store.
type MSSQLStore struct {
	db          *sql.DB
	stopCleanup chan bool
}

// New returns a new MSSQLStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *sql.DB) *MSSQLStore {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

// NewWithCleanupInterval returns a new MSSQLStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) *MSSQLStore {
	m := &MSSQLStore{db: db}
	if cleanupInterval > 0 {
		go m.startCleanup(cleanupInterval)
	}
	return m
}

// Find returns the data for a given session token from the MSSQLStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (m *MSSQLStore) Find(token string) (b []byte, exists bool, err error) {
	row := m.db.QueryRow("SELECT data FROM sessions WHERE token = @p1 AND GETUTCDATE() < expiry", token)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the MSSQLStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (m *MSSQLStore) Commit(token string, b []byte, expiry time.Time) error {
	_, err := m.db.Exec(`MERGE INTO sessions WITH (HOLDLOCK) AS T USING (VALUES(@p1)) AS S (token) ON (T.token = S.token)
						 WHEN MATCHED THEN UPDATE SET data = @p2, expiry = @p3
						 WHEN NOT MATCHED THEN INSERT (token, data, expiry) VALUES(@p1, @p2, @p3);`, token, b, expiry.UTC())
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the MSSQLStore
// instance.
func (m *MSSQLStore) Delete(token string) error {
	_, err := m.db.Exec("DELETE FROM sessions WHERE token = @p1", token)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the MSSQLStore instance.
func (m *MSSQLStore) All() (map[string][]byte, error) {
	rows, err := m.db.Query("SELECT token, data FROM sessions WHERE GETUTCDATE() < expiry")
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

func (m *MSSQLStore) startCleanup(interval time.Duration) {
	m.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := m.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-m.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the MSSQLStore
// instance. It's rare to terminate this; generally MSSQLStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the MSSQLStore is transient.
// An example is creating a new MSSQLStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// MSSQLStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (m *MSSQLStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *MSSQLStore) deleteExpired() error {
	_, err := m.db.Exec("DELETE FROM sessions WHERE expiry < GETUTCDATE()")
	return err
}
