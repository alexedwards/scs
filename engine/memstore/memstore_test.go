package memstore

import (
	"bytes"
	"testing"
	"time"

	"github.com/alexedwards/scs/session"
)

func TestNew(t *testing.T) {
	m := New()
	_, ok := interface{}(m).(session.Engine)
	if ok == false {
		t.Fatalf("got %v: expected %v", ok, true)
	}

	if len(m.Cache.Items()) > 0 {
		t.Fatalf("got %d: expected %d", len(m.Cache.Items()), 0)
	}
}

func TestFind(t *testing.T) {
	m := New()
	m.Cache.Set("test_session_token", []byte("encoded_data"), 0)

	b, found, err := m.Find("test_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}

	b, found, err = m.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if b != nil {
		t.Fatalf("got %v: expected %v", b, nil)
	}
}

func TestFindBadData(t *testing.T) {
	m := New()
	m.Cache.Set("test_session_token", "not_a_byte_slice", 0)

	_, _, err := m.Find("test_session_token")
	if err != ErrTypeAssertionFailed {
		t.Fatalf("got %v: expected %v", err, ErrTypeAssertionFailed)
	}
}

func TestExpiry(t *testing.T) {
	m := New()

	err := m.Save("test_session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	_, found, _ := m.Find("test_session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	time.Sleep(100 * time.Millisecond)
	_, found, _ = m.Find("test_session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSave(t *testing.T) {
	m := New()

	err := m.Save("test_session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if len(m.Cache.Items()) != 1 {
		t.Fatalf("got %d: expected %d", len(m.Cache.Items()), 1)
	}
}

func TestDelete(t *testing.T) {
	m := New()
	m.Cache.Set("test_session_token", []byte("encoded_data"), 0)

	err := m.Delete("test_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if len(m.Cache.Items()) != 0 {
		t.Fatalf("got %d: expected %d", len(m.Cache.Items()), 0)
	}
}
