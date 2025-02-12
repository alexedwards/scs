package sqlite3store

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createDBwithSessionTable(db *sql.DB) error {
	q := `CREATE TABLE sessions (
		token TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		expiry REAL NOT NULL
	);
	CREATE INDEX sessions_expiry_idx ON sessions(expiry);`
	_, err := db.Exec(q)
	if err != nil {
		return err
	}
	return nil
}

func removeDBfile(dsn string) error {
	fileinfo, _ := os.Stat(dsn)
	if fileinfo != nil {
		err := os.Remove(dsn)
		if err != err {
			return err
		}
	}
	return nil
}

func TestFind(t *testing.T) {
	dsn := "./testSQL3lite.db"

	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(dsn)
	defer db.Close()

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO sessions VALUES('session_token', 'encoded_data', datetime(current_timestamp, '+1 minute'))")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)

	b, found, err := p.Find("session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}
func TestFindMissing(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dsn)
	defer db.Close()

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO sessions VALUES('session_token', 'encoded_data', datetime(current_timestamp, '+1 minute'))")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)

	_, found, err := p.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dsn)
	defer db.Close()

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)

	err = p.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
	var data []byte
	err = row.Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dsn)
	defer db.Close()

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO sessions VALUES('session_token', 'encoded_data', datetime(current_timestamp, '+1 minute'))")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)

	err = p.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
	var data []byte
	err = row.Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove(dsn)

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)
	fmt.Print()
	err = p.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := p.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = p.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanup(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove(dsn)

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}

	if err := createDBwithSessionTable(db); err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 200*time.Millisecond)

	err = p.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(300 * time.Millisecond)
	p.StopCleanup()

	row = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	dsn := "./testSQL3lite.db"
	if err := removeDBfile(dsn); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove(dsn)

	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(db, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	p.StopCleanup()
}
