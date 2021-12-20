package leveldbstore

import (
	"bytes"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestCommit(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 0)
	err = ls.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := db.Get([]byte(basePrefix+"key1"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data[8:], []byte("value1")) {
		t.Fatalf("expected bytes `value1`, got %s", data)
	}
}

func TestFind(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 0)
	ls.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	data, found, err := ls.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if !bytes.Equal(data, []byte("value1")) {
		t.Fatalf("got %v: expected %v", data, []byte("value1"))
	}

	data, found, err = ls.Find("key2")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if data != nil {
		t.Fatalf("got %v, expected %v", data, nil)
	}
}

func TestDelete(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 0)

	err = ls.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := ls.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v, expected %v", found, true)
	}

	err = ls.Delete("key1")
	if err != nil {
		t.Fatal(err)
	}

	_, found, err = ls.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v, expected %v", found, false)
	}
}

func TestExpire(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 0)

	err = ls.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := ls.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = ls.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanup(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 10*time.Millisecond)
	err = ls.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	data, err := db.Get([]byte(basePrefix+"session_token"), nil)
	if err != leveldb.ErrNotFound {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("expected [], got %v", data)
	}

	ls.StopCleanup()
}

func TestStopNilCleanup(t *testing.T) {
	db, err := leveldb.OpenFile("/tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ls := NewWithCleanupInterval(db, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	ls.StopCleanup()
}
