// Package mysqlstore is a MySQL-based storage engine for the SCS session package.
//
// A working MySQL database is required, containing a sessions table with
// the definition:
//
//	CREATE TABLE sessions (
//	  token CHAR(43) PRIMARY KEY,
//	  data BLOB NOT NULL,
//	  expiry TIMESTAMP(6) NOT NULL
//	);
//	CREATE INDEX sessions_expiry_idx ON sessions (expiry);
//
// The mysqlstore package provides a background 'cleanup' goroutine to delete expired
// session data. This stops the database table from holding on to invalid sessions
// forever and growing unnecessarily large.
//
// Usage:
//
//  func main() {
//      // Establish a database/sql pool
//      db, err := sql.Open("mysql", "user:pass@/db")
//      if err != nil {
//          log.Fatal(err)
//      }
//      defer db.Close()
//
//      // Create a new MySQLStore instance using the existing database/sql pool,
//      // with a cleanup interval of 5 minutes.
//      engine := mysqlstore.New(db, 5*time.Minute)
//
//      sessionManager := session.Manage(engine)
//      http.ListenAndServe(":4000", sessionManager(http.DefaultServeMux))
//  }
//
// It is underpinned by the go-sql-driver/mysql driver (https://github.com/go-sql-driver/mysql).
package mysqlstore

import (
	"database/sql"
	"log"
	"time"

	// Register go-sql-driver/mysql with database/sql
	_ "github.com/go-sql-driver/mysql"
)

// MySQLStore represents the currently configured session storage engine.
type MySQLStore struct {
	*sql.DB
	stopCleanup chan bool
}

// New returns a new MySQLStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background cleanup goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(db *sql.DB, cleanupInterval time.Duration) *MySQLStore {
	m := &MySQLStore{DB: db}
	if cleanupInterval > 0 {
		go m.startCleanup(cleanupInterval)
	}
	return m
}

// Find returns the data for a given session token from the MySQLStore instance. If
// the session token is not found or is expired, the returned exists flag will be
// set to false.
func (m *MySQLStore) Find(token string) ([]byte, bool, error) {
	var b []byte
	row := m.DB.QueryRow("SELECT data FROM sessions WHERE token = ? AND UTC_TIMESTAMP(6) < expiry", token)
	err := row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Save adds a session token and data to the MySQLStore instance with the given expiry
// time. If the session token already exists then the data and expiry time are updated.
func (m *MySQLStore) Save(token string, b []byte, expiry time.Time) error {
	_, err := m.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data), expiry = VALUES(expiry)", token, b, expiry.UTC())
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the MySQLStore instance.
func (m *MySQLStore) Delete(token string) error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (m *MySQLStore) startCleanup(interval time.Duration) {
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

// StopCleanup terminates the background cleanup goroutine for the MySQLStore instance.
// It's rare to terminate this; generally MySQLStore instances and their cleanup
// goroutines are intended to be long-lived and run for the lifetime of  your
// application.
//
// There may be occasions though when your use of the MySQLStore is transient. An
// example is creating a new MySQLStore instance in a test function. In this scenario,
// the cleanup goroutine (which will run forever) will prevent the MySQLStore object
// from being garbage collected even after the test function has finished. You
// can prevent this by manually calling StopCleanup.
//
// Example:
//
//	func TestExample(t *testing.T) {
//		db, err := sql.Open("mysql", "user:pass@/db")
//		if err != nil {
//			t.Fatal(err)
//		}
//		defer db.Close()
//
//		engine := mysqlstore.New(db, time.Second)
//		defer engine.StopCleanup()
//
//		// Run test...
//	}
func (m *MySQLStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *MySQLStore) deleteExpired() error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE expiry < UTC_TIMESTAMP(6)")
	return err
}
