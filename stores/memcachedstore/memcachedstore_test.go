package memcachedstore

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func TestFind(t *testing.T) {
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	mc.Set(&memcache.Item{
		Key:        Prefix + "session_token",
		Value:      []byte("encoded_data"),
		Expiration: 60,
	})

	m := New(mc)

	b, found, err := m.Find("session_token")
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
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	m := New(mc)

	_, found, err := m.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	m := New(mc)

	err = m.Save("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	item, err := mc.Get(Prefix + "session_token")
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(item.Value, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", item.Value, []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	mc.Set(&memcache.Item{
		Key:        Prefix + "session_token",
		Value:      []byte("encoded_data"),
		Expiration: 60,
	})

	m := New(mc)

	err = m.Save("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	item, err := mc.Get(Prefix + "session_token")
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(item.Value, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", item.Value, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	m := New(mc)

	err = m.Save("session_token", []byte("encoded_data"), time.Now().Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := m.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)
	_, found, _ = m.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	mc := memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR"))

	err := mc.DeleteAll()
	if err != nil {
		t.Fatal(err)
	}

	mc.Set(&memcache.Item{
		Key:        Prefix + "session_token",
		Value:      []byte("encoded_data"),
		Expiration: 60,
	})

	m := New(mc)

	err = m.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = mc.Get(Prefix + "session_token")
	if err != memcache.ErrCacheMiss {
		t.Fatalf("got %v: expected %v", err, memcache.ErrCacheMiss)
	}
}
