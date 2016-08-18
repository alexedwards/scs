package pgstore

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PGStore struct {
	*sql.DB
	stopSweeper chan bool
}

func New(db *sql.DB, sweepInterval time.Duration) *PGStore {
	p := &PGStore{DB: db}
	if sweepInterval > 0 {
		go p.startSweeper(sweepInterval)
	}
	return p
}

func (p *PGStore) Find(token string) ([]byte, bool, error) {
	var b []byte
	row := p.DB.QueryRow("SELECT data FROM sessions WHERE token = $1 AND current_timestamp < expiry", token)
	err := row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (p *PGStore) Save(token string, b []byte, expiry time.Time) error {
	_, err := p.DB.Exec("INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, $3) ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry", token, b, expiry)
	if err != nil {
		return err
	}
	return nil
}

func (p *PGStore) Delete(token string) error {
	_, err := p.DB.Exec("DELETE FROM sessions WHERE token = $1", token)
	return err
}

func (p *PGStore) startSweeper(interval time.Duration) {
	p.stopSweeper = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := p.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-p.stopSweeper:
			ticker.Stop()
			return
		}
	}
}

func (p *PGStore) StopSweeper() {
	if p.stopSweeper != nil {
		p.stopSweeper <- true
	}
}

func (p *PGStore) deleteExpired() error {
	_, err := p.DB.Exec("DELETE FROM sessions WHERE expiry < current_timestamp")
	return err
}
