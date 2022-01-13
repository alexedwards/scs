package bunstore

import (
	"context"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// BunStore represents the session store.
type BunStore struct {
	db          *bun.DB
	stopCleanup chan bool
}

type session struct {
	bun.BaseModel `bun:"table:sessions,alias:sessions"`

	Token  string    `bun:"token,pk"`
	Data   []byte    `bun:"data"`
	Expiry time.Time `bun:"expiry"`
}

// New returns a new BunStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *bun.DB) (*BunStore, error) {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

// NewWithCleanupInterval returns a new BunStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *bun.DB, cleanupInterval time.Duration) (*BunStore, error) {
	b := &BunStore{db: db}

	if cleanupInterval > 0 {
		go b.startCleanup(cleanupInterval)
	}

	return b, nil
}

// Find returns the data for a given session token from the BunStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (b *BunStore) FindCtx(ctx context.Context, token string) (bb []byte, exists bool, err error) {
	s := &session{}
	count, err := b.db.NewSelect().Model(s).Where("token = ? AND expiry >= ?", token, time.Now()).ScanAndCount(ctx)
	if count == 0 {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return s.Data, true, nil
}

// Commit adds a session token and data to the BunStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (b *BunStore) CommitCtx(ctx context.Context, token string, bb []byte, expiry time.Time) error {
	s := &session{Token: token, Data: bb, Expiry: expiry}

	switch b.db.Dialect().Name() {
	case dialect.SQLite:
		if _, err := b.db.NewInsert().Model(s).Replace().Exec(ctx); err != nil {
			return err
		}
	case dialect.PG:
		if _, err := b.db.NewInsert().Model(s).On("CONFLICT (token) DO UPDATE").Set("data = EXCLUDED.data").Exec(ctx); err != nil {
			return err
		}
	case dialect.MySQL:
		if _, err := b.db.NewInsert().Model(s).On("DUPLICATE KEY UPDATE").Set("data = VALUES(data), expiry = VALUES(expiry)").Exec(ctx); err != nil {
			return err
		}
	default:
		panic("not reached")
	}

	return nil
}

// Delete removes a session token and corresponding data from the BunStore
// instance.
func (b *BunStore) DeleteCtx(ctx context.Context, token string) error {
	_, err := b.db.NewDelete().Model(&session{}).Where("token = ?", token).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the BunStore instance.
func (b *BunStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	rows, err := b.db.NewSelect().Model(&[]session{}).Where("expiry >= ?", time.Now()).Rows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ss := make(map[string][]byte)

	for rows.Next() {
		s := &session{}
		if err = b.db.ScanRow(ctx, rows, s); err != nil {
			return nil, err
		}

		ss[s.Token] = s.Data
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}

	return ss, nil
}

func (b *BunStore) startCleanup(interval time.Duration) {
	b.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := b.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-b.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the BunStore
// instance. It's rare to terminate this; generally BunStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the BunStore is transient.
// An example is creating a new BunStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// BunStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (b *BunStore) StopCleanup() {
	if b.stopCleanup != nil {
		b.stopCleanup <- true
	}
}

func (b *BunStore) deleteExpired() error {
	ctx := context.Background()
	_, err := b.db.NewDelete().Model(&session{}).Where("expiry < ?", time.Now()).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a seperate type makes any errors caught
// only at runtime instead of compile time. Oh well.

func (b *BunStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}
func (b *BunStore) Commit(token string, bb []byte, expiry time.Time) error {
	panic("missing context arg")
}
func (b *BunStore) Delete(token string) error {
	panic("missing context arg")
}
