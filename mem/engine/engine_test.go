package engine

import (
	"bytes"
	"testing"
	"time"

	"github.com/alexedwards/scs"
)

func TestNew(t *testing.T) {
	e := New()
	_, ok := interface{}(e).(scs.Engine)
	if ok == false {
		t.Fatalf("got %v: expected %v", ok, true)
	}

	if len(e.Cache.Items()) > 0 {
		t.Fatalf("got %d: expected %d", len(e.Cache.Items()), 0)
	}
}

func TestFind(t *testing.T) {
	e := New()
	e.Cache.Set("test_session_token", []byte("encoded_data"), 0)

	b, found, err := e.Find("test_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}

	b, found, err = e.Find("missing_session_token")
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
	e := New()
	e.Cache.Set("test_session_token", "not_a_byte_slice", 0)

	_, _, err := e.Find("test_session_token")
	if err != ErrTypeAssertionFailed {
		t.Fatalf("got %v: expected %v", err, ErrTypeAssertionFailed)
	}
}

func TestExpiry(t *testing.T) {
	e := New()

	err := e.Save("test_session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	_, found, _ := e.Find("test_session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	time.Sleep(100 * time.Millisecond)
	_, found, _ = e.Find("test_session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSave(t *testing.T) {
	e := New()

	err := e.Save("test_session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if len(e.Cache.Items()) != 1 {
		t.Fatalf("got %d: expected %d", len(e.Cache.Items()), 1)
	}
}

func TestDelete(t *testing.T) {
	e := New()
	e.Cache.Set("test_session_token", []byte("encoded_data"), 0)

	err := e.Delete("test_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if len(e.Cache.Items()) != 0 {
		t.Fatalf("got %d: expected %d", len(e.Cache.Items()), 0)
	}
}
