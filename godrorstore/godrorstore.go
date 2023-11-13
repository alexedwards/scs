package godrorstore

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type GodrorStore struct {
	db          *sql.DB
	stopCleanup chan bool
}

func New(db *sql.DB) *GodrorStore {
	return NewWithCleanupInterval(db, 5*time.Minute)
}

func NewWithCleanupInterval(db *sql.DB, cleanupInterval time.Duration) *GodrorStore {
	g := &GodrorStore{db: db}
	if cleanupInterval > 0 {
		go g.StartCleanup(cleanupInterval)
	}
	return g
}

func (g *GodrorStore) Find(token string) (b []byte, exists bool, err error) {
	stmt := fmt.Sprintf("SELECT data FROM sessions WHERE token = '%x' AND current_timestamp < expiry", token)
	row := g.db.QueryRow(stmt)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (g *GodrorStore) Commit(token string, b []byte, expiry time.Time) error {
	stmt := fmt.Sprintf("SELECT data FROM sessions WHERE token = '%x'", token)
	row := g.db.QueryRow(stmt)
	err := row.Err()
	if row.Scan() == sql.ErrNoRows {
		stmt = `INSERT INTO sessions (token, data, expiry) VALUES ('%x', '%x', to_timestamp('` + string(expiry.Format("2006-01-02 15:04:05.00")) + `', 'YYYY-MM-DD HH24:MI:SS.FF'))`
		stmt = fmt.Sprintf(stmt, token, b)
		fmt.Println(stmt)
		_, err := g.db.Exec(stmt)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	stmt = `UPDATE sessions SET data = '%x', expiry = to_timestamp('` + string(expiry.Format("2006-01-02 15:04:05.00")) + `', 'YYYY-MM-DD HH24:MI:SS.FF') WHERE token = '%x'`
	stmt = fmt.Sprintf(stmt, b, token)
	_, err = g.db.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}

func (g *GodrorStore) Delete(token string) error {
	stmt := fmt.Sprintf("DELETE FROM session WHERE token = '%x'", token)
	_, err := g.db.Exec(stmt)
	return err
}

func (g *GodrorStore) All() (map[string][]byte, error) {
	rows, err := g.db.Query("SELECT token, data FROM sessions WHERE current_timestamp < expiry")
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

func (g *GodrorStore) StartCleanup(interval time.Duration) {
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

func (g *GodrorStore) StopCleanup() {
	if g.stopCleanup != nil {
		g.stopCleanup <- true
	}
}

func (g *GodrorStore) deleteExpired() error {
	_, err := g.db.Exec("DELETE FROM sessions WHERE expiry < current_timestamp")
	return err
}
