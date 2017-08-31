package buntstore

import (
	"bytes"
	"os"
	"testing"
	"time"

	"strings"

	"github.com/tidwall/buntdb"
)

// remove old test DB if it exists and create a new one
func getTestDatabase() *buntdb.DB {
	err := os.Remove("/tmp/testing.db")
	if err != nil {
		panic(err)
	}
	db, err := buntdb.Open("/tmp/testing.db")
	if err != nil {
		panic(err)
	}
	return db
}

func TestSave(t *testing.T) {
	db := getTestDatabase()
	defer db.Close()

	bs := New(db)
	bs.Save("key1", []byte("value1"), time.Now().Add(time.Minute))

	db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get("key1")
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}
		if !strings.EqualFold(v, "value1") {
			t.Fatalf("expected string `value1`, got %s", v)
		}
		return nil
	})
}

func TestFind(t *testing.T) {
	db := getTestDatabase()
	defer db.Close()

	bs := New(db)
	bs.Save("key1", []byte("value1"), time.Now().Add(time.Minute))

	{
		v, found, err := bs.Find("key1")
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatalf("got %v: expected %v (%s)", found, true, v)
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
		if found {
			t.Fatalf("got %v: expected %v", found, false)
		}
		if v != nil {
			t.Fatalf("got %v, expected %v", v, nil)
		}
	}
}

func TestDelete(t *testing.T) {
	db := getTestDatabase()
	defer db.Close()

	bs := New(db)

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
		if !found {
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
		if found {
			t.Fatalf("got %v, expected %v", found, false)
		}
	}

}

func TestExpire(t *testing.T) {
	db := getTestDatabase()
	defer db.Close()

	bs := New(db)

	err := bs.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := bs.Find("session_token")
	if !found {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(10 * time.Millisecond)
	_, found, _ = bs.Find("session_token")
	if !found {
		t.Fatalf("got %v: expected %v", found, false)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = bs.Find("session_token")
	if found {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanup(t *testing.T) {
	db := getTestDatabase()
	defer db.Close()

	bs := New(db)
	err := bs.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	{
		err := db.View(func(tx *buntdb.Tx) error {
			data, err := tx.Get("session_token")
			if err == nil {
				t.Fatalf("expected not found, got %s", err.Error())
			}
			if data != "" {
				t.Fatalf("expected empty, got %v", data)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
