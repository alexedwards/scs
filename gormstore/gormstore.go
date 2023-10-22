package gormstore

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// GORMStore represents the session store.
type GORMStore struct {
	db          *gorm.DB
	stopCleanup chan bool
}

type session struct {
	Token  string    `gorm:"column:token;primaryKey;type:varchar(43)"`
	Data   []byte    `gorm:"column:data"`
	Expiry time.Time `gorm:"column:expiry;index"`
}

func (session) TableName() string {
	return "sessions"
}

// New returns a new GORMStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(db *gorm.DB) (*GORMStore, error) {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

// NewWithCleanupInterval returns a new GORMStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *gorm.DB, cleanupInterval time.Duration) (*GORMStore, error) {
	g := &GORMStore{db: db}
	if err := g.migrate(); err != nil {
		return nil, err
	}
	if cleanupInterval > 0 {
		go g.startCleanup(cleanupInterval)
	}
	return g, nil
}

// Find returns the data for a given session token from the GORMStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (g *GORMStore) Find(token string) (b []byte, exists bool, err error) {
	s := &session{}
	row := g.db.Where("token = ? AND expiry >= ?", token, time.Now()).Limit(1).Find(s)
	if row.Error != nil {
		return nil, false, row.Error
	}

	if row.RowsAffected == 0 {
		return nil, false, nil
	}

	return s.Data, true, nil
}

// Commit adds a session token and data to the GORMStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (g *GORMStore) Commit(token string, b []byte, expiry time.Time) error {
	s := &session{}
	row := g.db.Where(session{Token: token}).Assign(session{Data: b, Expiry: expiry}).FirstOrCreate(s)
	if row.Error != nil {
		return row.Error
	}
	return nil
}

// Delete removes a session token and corresponding data from the GORMStore
// instance.
func (g *GORMStore) Delete(token string) error {
	row := g.db.Delete(&session{}, "token = ?", token)
	if row.Error != nil {
		return row.Error
	}
	return nil
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the GORMStore instance.
func (g *GORMStore) All() (map[string][]byte, error) {
	rows, err := g.db.Find(&[]session{}, "expiry >= ?", time.Now()).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ss := make(map[string][]byte)
	for rows.Next() {
		s := &session{}
		err := g.db.ScanRows(rows, s)
		if err != nil {
			return nil, err
		}
		ss[s.Token] = s.Data
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ss, nil
}

func (g *GORMStore) migrate() error {
	var tableOptions string
	// Set table options for MySQL database dialect.
	if g.db.Dialector.Name() == "mysql" {
		tableOptions = "ENGINE=InnoDB CHARSET=utf8mb4"
	}
	err := g.db.Set("gorm:table_options", tableOptions).AutoMigrate(&session{})
	if err != nil {
		return err
	}
	return nil
}

func (g *GORMStore) startCleanup(interval time.Duration) {
	g.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := g.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-g.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the GORMStore
// instance. It's rare to terminate this; generally GORMStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the GORMStore is transient.
// An example is creating a new GORMStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// GORMStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (g *GORMStore) StopCleanup() {
	if g.stopCleanup != nil {
		g.stopCleanup <- true
	}
}

func (g *GORMStore) deleteExpired() error {
	row := g.db.Delete(&session{}, "expiry < ?", time.Now())
	if row.Error != nil {
		return row.Error
	}
	return nil
}
