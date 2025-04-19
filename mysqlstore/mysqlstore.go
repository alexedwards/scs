package mysqlstore

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// MySQLStore represents the session store.
type MySQLStore struct {
	*sql.DB
	version     string
	stopCleanup chan bool
	tableName   string
}

type Config struct {
	// CleanUpInterval is the interval between each cleanup operation.
	// If set to 0, the cleanup operation is disabled.
	CleanUpInterval time.Duration

	// TableName is the name of the table where the session data will be stored.
	// If not set, it will default to "sessions".
	TableName string
}

// New returns a new MySQLStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *sql.DB) *MySQLStore {
	return NewWithConfig(db, Config{
		CleanUpInterval: 5 * time.Minute,
	})
}

// NewWithCleanupInterval returns a new MySQLStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) *MySQLStore {
	return NewWithConfig(db, Config{
		CleanUpInterval: cleanupInterval,
	})
}

// NewWithConfig returns a new MySQLStore instance with the given configuration.
// If the TableName field is empty, it will be set to "sessions".
// If the CleanUpInterval field is 0, the cleanup goroutine will not be started.
func NewWithConfig(db *sql.DB, config Config) *MySQLStore {
	if config.TableName == "" {
		config.TableName = "sessions"
	}

	m := &MySQLStore{
		DB:        db,
		version:   getVersion(db),
		tableName: config.TableName,
	}

	if config.CleanUpInterval > 0 {
		go m.startCleanup(config.CleanUpInterval)
	}

	return m
}

// Find returns the data for a given session token from the MySQLStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (m *MySQLStore) Find(token string) ([]byte, bool, error) {
	var b []byte
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = fmt.Sprintf("SELECT data FROM %s WHERE token = ? AND UTC_TIMESTAMP(6) < expiry", m.tableName)
	} else {
		stmt = fmt.Sprintf("SELECT data FROM %s WHERE token = ? AND UTC_TIMESTAMP < expiry", m.tableName)
	}

	row := m.DB.QueryRow(stmt, token)
	err := row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the MySQLStore instance with the given
// expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (m *MySQLStore) Commit(token string, b []byte, expiry time.Time) error {
	stmt := fmt.Sprintf("INSERT INTO %s (token, data, expiry) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data), expiry = VALUES(expiry)", m.tableName)
	_, err := m.DB.Exec(stmt, token, b, expiry.UTC())
	return err
}

// Delete removes a session token and corresponding data from the MySQLStore
// instance.
func (m *MySQLStore) Delete(token string) error {
	stmt := fmt.Sprintf("DELETE FROM %s WHERE token = ?", m.tableName)
	_, err := m.DB.Exec(stmt, token)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the MySQLStore instance.
func (m *MySQLStore) All() (map[string][]byte, error) {
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = fmt.Sprintf("SELECT token, data FROM %s WHERE UTC_TIMESTAMP(6) < expiry", m.tableName)
	} else {
		stmt = fmt.Sprintf("SELECT token, data FROM %s WHERE UTC_TIMESTAMP < expiry", m.tableName)
	}

	rows, err := m.DB.Query(stmt)
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

// StopCleanup terminates the background cleanup goroutine for the MySQLStore
// instance. It's rare to terminate this; generally MySQLStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the MySQLStore is transient.
// An example is creating a new MySQLStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// MySQLStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (m *MySQLStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *MySQLStore) deleteExpired() error {
	var stmt string

	if compareVersion("5.6.4", m.version) >= 0 {
		stmt = fmt.Sprintf("DELETE FROM %s WHERE expiry < UTC_TIMESTAMP(6)", m.tableName)
	} else {
		stmt = fmt.Sprintf("DELETE FROM %s WHERE expiry < UTC_TIMESTAMP", m.tableName)
	}

	_, err := m.DB.Exec(stmt)
	return err
}

func getVersion(db *sql.DB) string {
	var version string
	row := db.QueryRow("SELECT VERSION()")
	err := row.Scan(&version)
	if err != nil {
		return ""
	}
	return strings.Split(version, "-")[0]
}

// Based on https://stackoverflow.com/a/26729704
func compareVersion(a, b string) (ret int) {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	loopMax := len(bs)
	if len(as) > len(bs) {
		loopMax = len(as)
	}
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(as) > i {
			x = as[i]
		}
		if len(bs) > i {
			y = bs[i]
		}
		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			ret = -1
		} else if xi < yi {
			ret = 1
		}
		if ret != 0 {
			break
		}
	}
	return
}
