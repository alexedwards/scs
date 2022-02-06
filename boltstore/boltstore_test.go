package boltstore

import (
	"bytes"
	"testing"
	"time"

	"go.etcd.io/bbolt"
)

func TestCommit(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bs := NewWithCleanupInterval(db, 0)
	bs.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))

	db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		v := bucket.Get([]byte("key1"))
		if !bytes.Equal(v[8:], []byte("value1")) {
			t.Fatalf("expected bytes `value1`, got %s", v)
		}
		return nil
	})
}

func TestFind(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bs := NewWithCleanupInterval(db, 0)
	bs.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	v, found, err := bs.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if !bytes.Equal(v, []byte("value1")) {
		t.Fatalf("got %v: expected %v", v, []byte("value1"))
	}

	v, found, err = bs.Find("key2")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if v != nil {
		t.Fatalf("got %v, expected %v", v, nil)
	}
}

func TestDelete(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bs := NewWithCleanupInterval(db, 0)

	err = bs.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := bs.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v, expected %v", found, true)
	}

	err = bs.Delete("key1")
	if err != nil {
		t.Fatal(err)
	}

	_, found, err = bs.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v, expected %v", found, false)
	}
}

func TestExpire(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bs := NewWithCleanupInterval(db, 0)

	err = bs.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := bs.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = bs.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanup(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bs := NewWithCleanupInterval(db, 10*time.Millisecond)
	err = bs.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		data := bucket.Get([]byte("session_token"))
		if data != nil {
			t.Fatalf("expected nil, got %v", data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	bs.StopCleanup()
}

func TestStopNilCleanup(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	m := NewWithCleanupInterval(db, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	m.StopCleanup()
}

func TestDeleteExpired(t *testing.T) {
	db, err := bbolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	m := NewWithCleanupInterval(db, 0)

	if err := m.Commit("session_token1", []byte("data"), time.Now().Add(10*time.Millisecond)); err != nil {
		t.Fatal(err)
	}

	if err := m.Commit("session_token2", []byte("data"), time.Now().Add(10*time.Millisecond)); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := m.deleteExpired(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		data := bucket.Get([]byte("session_token1"))
		if data != nil {
			t.Fatalf("expected nil, got %v", data)
		}

		data = bucket.Get([]byte("session_token2"))
		if data != nil {
			t.Fatalf("expected nil, got %v", data)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
