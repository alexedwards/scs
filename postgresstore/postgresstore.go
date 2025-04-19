package postgresstore

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Config struct {
	// CleanUpInterval is the interval between each cleanup operation.
	// If set to 0, the cleanup operation is disabled.
	CleanUpInterval time.Duration

	// TableName is the name of the table where the session data will be stored.
	// If not set, it will default to "sessions".
	TableName string
}

// PostgresStore represents the session store.
type PostgresStore struct {
	db          *sql.DB
	stopCleanup chan bool
	tableName   string
}

// New returns a new PostgresStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *sql.DB) *PostgresStore {
	return NewWithConfig(db, Config{CleanUpInterval: 5 * time.Minute, TableName: "sessions"})
}

// NewWithCleanupInterval returns a new PostgresStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) *PostgresStore {
	return NewWithConfig(db, Config{CleanUpInterval: cleanupInterval, TableName: "sessions"})
}

// NewWithConfig returns a new PostgresStore instance using the provided config.
// If the TableName field is empty, it will be set to "sessions".
// If the CleanUpInterval field is 0, the cleanup goroutine will not be started.
func NewWithConfig(db *sql.DB, config Config) *PostgresStore {
	if config.TableName == "" {
		config.TableName = "sessions"
	}

	store := &PostgresStore{
		db:        db,
		tableName: config.TableName,
	}

	if config.CleanUpInterval > 0 {
		go store.startCleanup(config.CleanUpInterval)
	}

	return store
}

// Find returns the data for a given session token from the PostgresStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *PostgresStore) Find(token string) (b []byte, exists bool, err error) {
	row := p.db.QueryRow(
		fmt.Sprintf("SELECT data FROM %s WHERE token = $1 AND current_timestamp < expiry", p.tableName),
		token,
	)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the PostgresStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *PostgresStore) Commit(token string, b []byte, expiry time.Time) error {
	_, err := p.db.Exec(
		fmt.Sprintf("INSERT INTO %s (token, data, expiry) VALUES ($1, $2, $3) ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry", p.tableName),
		token,
		b,
		expiry,
	)
	return err
}

// Delete removes a session token and corresponding data from the PostgresStore
// instance.
func (p *PostgresStore) Delete(token string) error {
	_, err := p.db.Exec(fmt.Sprintf(
		"DELETE FROM %s WHERE token = $1",
		p.tableName),
		token,
	)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the PostgresStore instance.
func (p *PostgresStore) All() (map[string][]byte, error) {
	rows, err := p.db.Query(fmt.Sprintf("SELECT token, data FROM %s WHERE current_timestamp < expiry", p.tableName))
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

func (p *PostgresStore) startCleanup(interval time.Duration) {
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

// StopCleanup terminates the background cleanup goroutine for the PostgresStore
// instance. It's rare to terminate this; generally PostgresStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the PostgresStore is transient.
// An example is creating a new PostgresStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// PostgresStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (p *PostgresStore) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *PostgresStore) deleteExpired() error {
	_, err := p.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE expiry < current_timestamp", p.tableName))
	return err
}
