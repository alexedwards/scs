package boltstore

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

func TestSave(t *testing.T) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bs := New(db, time.Minute)
	bs.Save("key1", []byte("value1"), time.Now().Add(time.Minute))

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(dataBucketName)
		v := bucket.Get([]byte("key1"))
		if !bytes.Equal(v, []byte("value1")) {
			t.Fatalf("expected bytes `value1`, got %s", v)
		}
		return nil
	})
}

func TestFind(t *testing.T) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bs := New(db, time.Minute)
	bs.Save("key1", []byte("value1"), time.Now().Add(time.Minute))

	{
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
	}

	{
		v, found, err := bs.Find("key2")
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
}

func TestDelete(t *testing.T) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bs := New(db, time.Minute)

	{
		err := bs.Save("key1", []byte("value1"), time.Now().Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		_, found, err := bs.Find("key1")
		if err != nil {
			t.Fatal(err)
		}
		if found != true {
			t.Fatalf("got %v, expected %v", found, true)
		}
	}

	{
		err := bs.Delete("key1")
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		_, found, err := bs.Find("key1")
		if err != nil {
			t.Fatal(err)
		}
		if found != false {
			t.Fatalf("got %v, expected %v", found, false)
		}
	}

}

func TestExpire(t *testing.T) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bs := New(db, time.Minute)

	err = bs.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
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
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bs := New(db, time.Millisecond*10)
	err = bs.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	{
		err := db.View(func(tx *bolt.Tx) error {
			dataBucket := tx.Bucket(dataBucketName)
			expiryBucket := tx.Bucket(expiryBucketName)
			data := dataBucket.Get([]byte("session_token"))
			if data != nil {
				t.Fatalf("expected nil, got %v", data)
			}
			exp := expiryBucket.Get([]byte("session_token"))
			if exp != nil {
				t.Fatalf("expected nil, got %v", exp)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	bs.StopCleanup()
}

func TestStopNilCleanup(t *testing.T) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := New(db, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	m.StopCleanup()
}
