package session

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/alexedwards/scs/mem/engine"
)

func TestString(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutString", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, cookie = testRequest(t, h, "/PopString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}
}

func TestBool(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutBool", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetBool", cookie)
	if body != "true" {
		t.Fatalf("got %q: expected %q", body, "true")
	}

	_, body, cookie = testRequest(t, h, "/PopBool", cookie)
	if body != "true" {
		t.Fatalf("got %q: expected %q", body, "true")
	}

	_, body, _ = testRequest(t, h, "/GetBool", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}
}

func TestInt(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutInt", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetInt", cookie)
	if body != "12345" {
		t.Fatalf("got %q: expected %q", body, "12345")
	}

	_, body, cookie = testRequest(t, h, "/PopInt", cookie)
	if body != "12345" {
		t.Fatalf("got %q: expected %q", body, "12345")
	}

	_, body, _ = testRequest(t, h, "/GetInt", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	r := requestWithSession(new(http.Request), &session{data: make(map[string]interface{})})

	_ = PutInt(r, "test_int", 12345)
	i, _ := GetInt(r, "test_int")
	if i != 12345 {
		t.Fatalf("got %d: expected %d", i, 12345)
	}
}

func TestFloat(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutFloat", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetFloat", cookie)
	if body != "12.345" {
		t.Fatalf("got %q: expected %q", body, "12.345")
	}

	_, body, cookie = testRequest(t, h, "/PopFloat", cookie)
	if body != "12.345" {
		t.Fatalf("got %q: expected %q", body, "12.345")
	}

	_, body, _ = testRequest(t, h, "/GetFloat", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	r := requestWithSession(new(http.Request), &session{data: make(map[string]interface{})})

	_ = PutFloat(r, "test_float", 12.345)
	i, _ := GetFloat(r, "test_float")
	if i != 12.345 {
		t.Fatalf("got %d: expected %d", i, 12.345)
	}
}

func TestTime(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutTime", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetTime", cookie)
	if body != "10 Nov 09 23:00 UTC" {
		t.Fatalf("got %q: expected %q", body, "10 Nov 09 23:00 UTC")
	}

	_, body, cookie = testRequest(t, h, "/PopTime", cookie)
	if body != "10 Nov 09 23:00 UTC" {
		t.Fatalf("got %q: expected %q", body, "10 Nov 09 23:00 UTC")
	}

	_, body, _ = testRequest(t, h, "/GetTime", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	r := requestWithSession(new(http.Request), &session{data: make(map[string]interface{})})

	tt := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	_ = PutTime(r, "test_time", tt)
	tm, _ := GetTime(r, "test_time")
	if tm != tt {
		t.Fatalf("got %v: expected %v", t, tt)
	}
}

func TestBytes(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, body, cookie := testRequest(t, h, "/PutBytes", "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetBytes", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, cookie = testRequest(t, h, "/PopBytes", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, _ = testRequest(t, h, "/GetBytes", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	r := requestWithSession(new(http.Request), &session{data: make(map[string]interface{})})

	_ = PutBytes(r, "test_bytes", []byte("lorem ipsum"))
	b, _ := GetBytes(r, "test_bytes")
	if bytes.Equal(b, []byte("lorem ipsum")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("lorem ipsum"))
	}

	err := PutBytes(r, "test_bytes", nil)
	if err == nil {
		t.Fatalf("expected an error")
	}
}

func TestRemove(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	_, _, cookie = testRequest(t, h, "/PutBool", cookie)

	_, body, cookie := testRequest(t, h, "/RemoveString", cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	_, body, _ = testRequest(t, h, "/GetBool", cookie)
	if body != "true" {
		t.Fatalf("got %q: expected %q", body, "true")
	}
}

func TestClear(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	_, _, cookie = testRequest(t, h, "/PutBool", cookie)

	_, body, cookie := testRequest(t, h, "/Clear", cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}

	_, body, _ = testRequest(t, h, "/GetBool", cookie)
	if body != ErrKeyNotFound.Error() {
		t.Fatalf("got %q: expected %q", body, ErrKeyNotFound.Error())
	}
}
