package qlstore

import (
	"bytes"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestFind(t *testing.T) {
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)
	_, err = execTx(db, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}
	ex := time.Now().Add(time.Minute)
	_, err = execTx(db,
		`INSERT INTO sessions VALUES("session_token", $1,$2 )`,
		[]byte("encoded_data"), ex)
	if err != nil {
		t.Fatal(err)
	}

	p := New(db, 0)

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
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	p := New(db, 0)

	_, found, err := p.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
func TestSaveNew(t *testing.T) {
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	p := New(db, 0)

	err = p.Save("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT data FROM sessions WHERE token = "session_token"`)
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
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	ex := time.Now().Add(time.Minute)
	_, err = execTx(db,
		`INSERT INTO sessions VALUES("session_token", $1,$2 )`,
		[]byte("encoded_data"), ex)
	if err != nil {
		t.Fatal(err)
	}
	p := New(db, 0)

	err = p.Save("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT data FROM sessions WHERE token = "session_token"`)
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
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	p := New(db, 0)

	err = p.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
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

func TestDelete(t *testing.T) {
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	ex := time.Now().Add(time.Minute)
	_, err = execTx(db,
		`INSERT INTO sessions VALUES("session_token", $1,$2 )`,
		[]byte("encoded_data"), ex)
	if err != nil {
		t.Fatal(err)
	}

	p := New(db, 0)

	err = p.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT count(*) FROM sessions WHERE token = "session_token"`)
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestCleanup(t *testing.T) {
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)
	_, err = execTx(db, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := New(db, 200*time.Millisecond)
	defer p.StopCleanup()

	err = p.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT count(*) FROM sessions WHERE token = "session_token"`)
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(300 * time.Millisecond)
	row = db.QueryRow(`SELECT count(*) FROM sessions WHERE token = "session_token"`)
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	dsn := os.Getenv("SESSION_QL_TEST_DSN")
	if dsn == "" {
		dsn = "test.db"
	}
	db, err := sql.Open("ql-mem", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	migrate(t, db)

	p := New(db, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	p.StopCleanup()
}

func migrate(t *testing.T, db *sql.DB) {
	_, err := execTx(db, Table())
	if err != nil {
		t.Error(err)
	}
}
