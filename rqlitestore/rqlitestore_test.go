package rqlitestore

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rqlite/gorqlite"
)

func openConnection() (*gorqlite.Connection, error) {
	dsn := os.Getenv("SCS_RQLITE_TEST_DSN")
	conn, err := gorqlite.Open(dsn)
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteOne("DROP TABLE IF EXISTS sessions")
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteOne("CREATE TABLE sessions (token TEXT PRIMARY KEY, data BLOB NOT NULL, expiry REAL NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteOne("CREATE INDEX sessions_expiry_idx ON sessions(expiry)")
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

func TestFind(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%x', datetime(current_timestamp, '+1 minute'))", "session_token", "encoded_data")
	_, err = conn.WriteOne(query)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	b, found, err := r.Find("session_token")
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
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%x', datetime(current_timestamp, '+1 minute'))", "session_token", "encoded_data")
	_, err = conn.WriteOne(query)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	_, found, err := r.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	err = r.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row, err := conn.QueryOne("SELECT data FROM sessions WHERE token = 'session_token'")
	if err != nil {
		t.Fatal(err)
	}

	var data []byte

	for row.Next() {
		var datax string

		err := row.Scan(&datax)
		if err != nil {
			t.Fatal(err)
		}

		data, err = hex.DecodeString(datax)
		if err != nil {
			t.Fatal(err)
		}
	}

	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%x', datetime(current_timestamp, '+1 minute'))", "session_token", "encoded_data")
	_, err = conn.WriteOne(query)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	err = r.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row, err := conn.QueryOne("SELECT data FROM sessions WHERE token = 'session_token'")
	if err != nil {
		t.Fatal(err)
	}

	var data []byte

	for row.Next() {
		var datax string

		err := row.Scan(&datax)
		if err != nil {
			t.Fatal(err)
		}

		data, err = hex.DecodeString(datax)
		if err != nil {
			t.Fatal(err)
		}
	}

	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	err = r.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := r.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)

	_, found, _ = r.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	query := fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%x', datetime(current_timestamp, '+1 minute'))", "session_token", "encoded_data")
	_, err = conn.WriteOne(query)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	err = r.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	row, err := conn.QueryOne("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	if err != nil {
		t.Fatal(err)
	}

	var count int64

	for row.Next() {
		err := row.Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
	}

	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestAll(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	log.Println(err)
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 0)

	setSessions := make(map[string][]byte)
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("session_token_%v", i)
		val := []byte(fmt.Sprintf("encoded_data_%v", i))

		query := fmt.Sprintf("INSERT INTO sessions VALUES('%s', '%x', datetime(current_timestamp, '+1 minute'))", key, val)
		_, err = conn.WriteOne(query)
		if err != nil {
			t.Fatal(err)
		}

		setSessions[key] = val
	}

	gotSessions, err := r.All()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(setSessions, gotSessions) == false {
		t.Fatalf("got %v: expected %v", gotSessions, setSessions)
	}
}

func TestCleanup(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.WriteOne("DELETE FROM sessions")
	if err != nil {
		t.Fatal(err)
	}

	r := NewWithCleanupInterval(*conn, 2*time.Second)
	defer r.StopCleanup()

	err = r.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	row, err := conn.QueryOne("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	if err != nil {
		t.Fatal(err)
	}

	var count int64

	for row.Next() {
		err := row.Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
	}

	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(3 * time.Second)

	row, err = conn.QueryOne("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	if err != nil {
		t.Fatal(err)
	}

	for row.Next() {
		err := row.Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
	}

	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	conn, err := openConnection()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	r := NewWithCleanupInterval(*conn, 0)
	time.Sleep(1 * time.Second)
	// A send to a nil channel will block forever
	r.StopCleanup()
}
