package memcachedstore

import (
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func TestFind(t *testing.T) {
	store := New(memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR")))

	store.Save("session_token", []byte("encoded_data"), time.Now().AddDate(0, 0, 5))
	val, found, err := store.Find("session_token")

	if string(val) != "encoded_data" {
		t.Errorf("Expected \"encoded_data\", got %q", val)
	}

	if !found {
		t.Errorf("Expected found to be true, but got %t", found)
	}

	if err != nil {
		t.Errorf("Encountered error %+v", err)
	}

	store.Delete("session_token")
}

func TestDelete(t *testing.T) {
	store := New(memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR")))

	store.Save("session_token", []byte("encoded_data"), time.Now().AddDate(0, 2, 0))
	store.Delete("session_token")
	val, found, err := store.Find("session_token")

	if val != nil {
		t.Errorf("Expected nil, got %q", val)
	}

	if found {
		t.Errorf("Expected found to be false, but got %t", found)
	}

	if err != memcache.ErrCacheMiss {
		t.Errorf("Delete did not cause cache miss error: %+v", err)
	}
}

func TestExpiry(t *testing.T) {
	store := New(memcache.New(os.Getenv("SESSION_MEMCACHED_TEST_ADDR")))

	exp := time.Duration(time.Second * 2)	// memcached is fickle if you test at the bare minimum of 1 second

	store.Save("session_token", []byte("encoded_data"), time.Now().Add(exp))
	time.Sleep(exp)
	val, found, err := store.Find("session_token")

	if val != nil {
		t.Errorf("Expected nil, got %q", val)
	}

	if found {
		t.Errorf("Expected found to be false, but got %t", found)
	}

	if err != memcache.ErrCacheMiss {
		t.Errorf("Expiration did not cause cache miss error: %+v", err)
	}
}
