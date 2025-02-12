package pgxstore

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore represents the session store.
type PostgresStore struct {
	pool        *pgxpool.Pool
	stopCleanup chan bool
}

// New returns a new PostgresStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(pool *pgxpool.Pool) *PostgresStore {
	return NewWithCleanupInterval(pool, 5*time.Minute)
}

// NewWithCleanupInterval returns a new PostgresStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(pool *pgxpool.Pool, cleanupInterval time.Duration) *PostgresStore {
	p := &PostgresStore{pool: pool}
	if cleanupInterval > 0 {
		p.stopCleanup = make(chan bool)
		go p.startCleanup(cleanupInterval)
	}
	return p
}

// FindCtx returns the data for a given session token from the PostgresStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *PostgresStore) FindCtx(ctx context.Context, token string) (b []byte, found bool, err error) {
	row := p.pool.QueryRow(ctx, "SELECT data FROM sessions WHERE token = $1 AND current_timestamp < expiry", token)
	err = row.Scan(&b)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// CommitCtx adds a session token and data to the PostgresStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *PostgresStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error) {
	_, err = p.pool.Exec(ctx, "INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, $3) ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry", token, b, expiry)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCtx removes a session token and corresponding data from the PostgresStore
// instance.
func (p *PostgresStore) DeleteCtx(ctx context.Context, token string) (err error) {
	_, err = p.pool.Exec(ctx, "DELETE FROM sessions WHERE token = $1", token)
	return err
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the PostgresStore instance.
func (p *PostgresStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	rows, err := p.pool.Query(ctx, "SELECT token, data FROM sessions WHERE current_timestamp < expiry")
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
	_, err := p.pool.Exec(context.Background(), "DELETE FROM sessions WHERE expiry < current_timestamp")
	return err
}

// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a separate type makes any errors caught
// only at runtime instead of compile time.

func (p *PostgresStore) Find(token string) (b []byte, exists bool, err error) {
	panic("missing context arg")
}

func (p *PostgresStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}

func (p *PostgresStore) Delete(token string) error {
	panic("missing context arg")
}

func (p *PostgresStore) All() (map[string][]byte, error) {
	panic("missing context arg")
}
