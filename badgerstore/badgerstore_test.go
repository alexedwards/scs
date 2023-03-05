package badgerstore

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
)

var (
	db *badger.DB
)

func TestMain(m *testing.M) {
	var (
		result int
		err    error
	)

	db, err = badger.Open(badger.DefaultOptions("test.db"))
	if err != nil {
		log.Fatal(err)
	}

	if err == nil {
		result = m.Run()
	}

	db.Close()
	err = os.RemoveAll("test.db")
	if err != nil {
		log.Println("Could not delete test store folder \"test.db\".")
		log.Println("You can delete it manually instead.")
	}
	os.Exit(result)
}

func TestFind(t *testing.T) {
	store := New(db)

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(store.prefix+"session_token"), []byte("encoded_data"))
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	data, found, err := store.Find("session_token")
	if err != nil {
		t.Fatal(err)
	}

	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	if bytes.Equal(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestSaveNew(t *testing.T) {
	store := New(db)

	err := store.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(store.prefix + "session_token"))
		if err != nil {
			log.Fatal(err)
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			log.Fatal(err)
		}

		if reflect.DeepEqual(data, []byte("encoded_data")) == false {
			t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestFindMissing(t *testing.T) {
	store := New(db)

	_, found, err := store.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveUpdated(t *testing.T) {
	store := New(db)

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(store.prefix+"session_token"), []byte("encoded_data"))
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	var data []byte

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(store.prefix + "session_token"))
		if err != nil {
			t.Fatal(err)
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	store := New(db)
	expiry := time.Now().Add(time.Second)

	err := store.Commit("session_token", []byte("encoded_data"), expiry)
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := store.Find("session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)

	_, found, err = store.Find("session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	store := New(db)

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(store.prefix+"session_token"), []byte("encoded_data"))
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(store.prefix + "session_token"))
		if err == badger.ErrKeyNotFound {
			return nil
		} else if err != nil {
			return err
		}
		t.Fatalf("got %v: expected %v", item, nil)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
