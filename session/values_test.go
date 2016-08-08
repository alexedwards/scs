package session

import (
	"testing"

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
