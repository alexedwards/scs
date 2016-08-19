package mysqlstore

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	*sql.DB
	stopSweeper chan bool
}

func New(db *sql.DB, sweepInterval time.Duration) *MySQLStore {
	m := &MySQLStore{DB: db}
	if sweepInterval > 0 {
		go m.startSweeper(sweepInterval)
	}
	return m
}

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

func (m *MySQLStore) Save(token string, b []byte, expiry time.Time) error {
	_, err := m.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data), expiry = VALUES(expiry)", token, b, expiry.UTC())
	if err != nil {
		return err
	}
	return nil
}

func (m *MySQLStore) Delete(token string) error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (m *MySQLStore) startSweeper(interval time.Duration) {
	m.stopSweeper = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := m.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-m.stopSweeper:
			ticker.Stop()
			return
		}
	}
}

func (m *MySQLStore) StopSweeper() {
	if m.stopSweeper != nil {
		m.stopSweeper <- true
	}
}

func (m *MySQLStore) deleteExpired() error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE expiry < UTC_TIMESTAMP(6)")
	return err
}
