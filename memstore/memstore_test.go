package memstore

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestFind(t *testing.T) {
	m := NewWithCleanupInterval(0)
	m.items["session_token"] = item{object: []byte("encoded_data"), expiration: time.Now().Add(time.Second).UnixNano()}

	b, found, err := m.Find("session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	m := NewWithCleanupInterval(0)

	_, found, err := m.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCommitNew(t *testing.T) {
	m := NewWithCleanupInterval(0)

	err := m.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	v, found := m.items["session_token"]
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	if reflect.DeepEqual(v.object, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", v.object, []byte("encoded_data"))
	}
}

func TestCommitUpdated(t *testing.T) {
	m := NewWithCleanupInterval(0)

	err := m.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	err = m.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	v := m.items["session_token"].object

	if reflect.DeepEqual(v, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", v, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	m := NewWithCleanupInterval(0)

	err := m.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	_, found, _ := m.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(101 * time.Millisecond)
	_, found, _ = m.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	m := NewWithCleanupInterval(0)
	m.items["session_token"] = item{object: []byte("encoded_data"), expiration: time.Now().Add(time.Second).UnixNano()}

	err := m.Delete("session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	_, found := m.items["session_token"]
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanupInterval(t *testing.T) {
	m := NewWithCleanupInterval(100 * time.Millisecond)
	defer m.StopCleanup()
	m.mu.Lock()
	m.items["session_token"] = item{object: []byte("encoded_data"), expiration: time.Now().Add(500 * time.Millisecond).UnixNano()}
	m.mu.Unlock()

	m.mu.Lock()
	_, ok := m.items["session_token"]
	m.mu.Unlock()
	if !ok {
		t.Fatalf("got %v: expected %v", ok, true)
	}

	time.Sleep(time.Second)
	m.mu.Lock()
	_, ok = m.items["session_token"]
	m.mu.Unlock()
	if ok {
		t.Fatalf("got %v: expected %v", ok, false)
	}
}
