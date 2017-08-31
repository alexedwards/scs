package memstore

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestFind(t *testing.T) {
	m := New(time.Minute)
	m.Cache.Set("session_token", []byte("encoded_data"), 0)

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
	m := New(time.Minute)

	_, found, err := m.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestFindBadData(t *testing.T) {
	m := New(time.Minute)
	m.Cache.Set("session_token", "not_a_byte_slice", 0)

	_, _, err := m.Find("session_token")
	if err != errTypeAssertionFailed {
		t.Fatalf("got %v: expected %v", err, errTypeAssertionFailed)
	}
}

func TestSaveNew(t *testing.T) {
	m := New(time.Minute)

	err := m.Save("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	v, found := m.Cache.Get("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	b, ok := v.([]byte)
	if ok == false {
		t.Fatal("could not convert to []byte")
	}
	if reflect.DeepEqual(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	m := New(time.Minute)
	m.Cache.Set("session_token", []byte("encoded_data"), 0)

	err := m.Save("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	err = m.Save("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	v, _ := m.Cache.Get("session_token")
	b, ok := v.([]byte)
	if ok == false {
		t.Fatal("could not convert to []byte")
	}
	if reflect.DeepEqual(b, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	m := New(time.Minute)

	err := m.Save("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	_, found, _ := m.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = m.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	m := New(time.Minute)
	m.Cache.Set("session_token", []byte("encoded_data"), 0)

	err := m.Delete("session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	_, found := m.Cache.Get("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
