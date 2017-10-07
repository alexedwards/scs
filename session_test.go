package scs

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

type testUser struct {
	Name string
	Age  int
}

func init() {
	gob.Register(new(testUser))
}

func TestGenerateToken(t *testing.T) {
	id, err := generateToken()
	if err != nil {
		t.Fatal(err)
	}

	match, err := regexp.MatchString("^[0-9a-zA-Z_\\-]{43}$", id)
	if err != nil {
		t.Fatal(err)
	}
	if match == false {
		t.Errorf("got %q: should match %q", id, "^[0-9a-zA-Z_\\-]{43}$")
	}
}

func TestString(t *testing.T) {
	manager := NewManager(newMockStore())

	_, body, cookie := testRequest(t, testPutString(manager), "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, testGetString(manager), cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, cookie = testRequest(t, testPopString(manager), cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, _ = testRequest(t, testGetString(manager), cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
}

func TestObject(t *testing.T) {
	manager := NewManager(newMockStore())

	_, body, cookie := testRequest(t, testPutObject(manager), "")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, testGetObject(manager), cookie)
	if body != "alice: 21" {
		t.Fatalf("got %q: expected %q", body, "alice: 21")
	}

	_, body, cookie = testRequest(t, testPopObject(manager), cookie)
	if body != "alice: 21" {
		t.Fatalf("got %q: expected %q", body, "alice: 21")
	}

	_, body, _ = testRequest(t, testGetObject(manager), cookie)
	if body != ": 0" {
		t.Fatalf("got %q: expected %q", body, ": 0")
	}
}

func TestDestroy(t *testing.T) {
	store := newMockStore()
	manager := NewManager(store)

	_, _, cookie := testRequest(t, testPutString(manager), "")
	oldToken := extractTokenFromCookie(cookie)

	_, body, cookie := testRequest(t, testDestroy(manager), cookie)

	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	if strings.HasPrefix(cookie, fmt.Sprintf("%s=;", manager.opts.name)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=;", manager.opts.name))
	}
	if strings.Contains(cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}
	if strings.Contains(cookie, "Max-Age=0") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Max-Age=0")
	}
	_, found, _ := store.Find(oldToken)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestRenewToken(t *testing.T) {
	store := newMockStore()
	manager := NewManager(store)

	_, _, cookie := testRequest(t, testPutString(manager), "")
	oldToken := extractTokenFromCookie(cookie)

	_, body, cookie := testRequest(t, testRenewToken(manager), cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatal("expected a difference")
	}
	_, found, _ := store.Find(oldToken)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}

	_, body, _ = testRequest(t, testGetString(manager), cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}
}

func TestClear(t *testing.T) {
	manager := NewManager(newMockStore())

	_, _, cookie := testRequest(t, testPutString(manager), "")
	_, _, cookie = testRequest(t, testPutBool(manager), cookie)

	_, body, cookie := testRequest(t, testClear(manager), cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}

	_, body, _ = testRequest(t, testGetString(manager), cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}

	_, body, _ = testRequest(t, testGetBool(manager), cookie)
	if body != "false" {
		t.Fatalf("got %q: expected %q", body, "false")
	}

	// Check that it's a no-op if there is no data in the session
	_, _, cookie = testRequest(t, testClear(manager), cookie)
	if cookie != "" {
		t.Fatalf("got %q: expected %q", cookie, "")
	}
}

func TestKeys(t *testing.T) {
	manager := NewManager(newMockStore())

	_, _, cookie := testRequest(t, testPutString(manager), "")
	_, _, _ = testRequest(t, testPutBool(manager), cookie)

	_, body, _ := testRequest(t, testKeys(manager), cookie)
	if body != "[test_bool test_string]" {
		t.Fatalf("got %q: expected %q", body, "[test_bool test_string]")
	}

	_, _, _ = testRequest(t, testClear(manager), cookie)
	_, body, _ = testRequest(t, testKeys(manager), cookie)
	if body != "[]" {
		t.Fatalf("got %q: expected %q", body, "[]")
	}
}

func TestLoadFailure(t *testing.T) {
	manager := NewManager(newMockStore())

	cookie := http.Cookie{
		Name:  "session",
		Value: "force-error",
	}

	_, body, _ := testRequest(t, testPutString(manager), cookie.String())
	if body != "forced-error\n" {
		t.Fatalf("got %q: expected %q", body, "forced-error\n")
	}
}

func TestMultipleSessions(t *testing.T) {
	manager1 := NewManager(newMockStore())
	manager1.Name("foo")

	_, _, cookie1 := testRequest(t, testPutString(manager1), "")

	manager2 := NewManager(newMockStore())
	manager2.Name("bar")

	_, _, cookie2 := testRequest(t, testPutBool(manager2), "")

	_, body, _ := testRequest(t, testGetString(manager1), cookie1)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	_, body, _ = testRequest(t, testGetBool(manager2), cookie2)
	if body != "true" {
		t.Fatalf("got %q: expected %q", body, "true")
	}

	_, body, _ = testRequest(t, testGetString(manager2), cookie2)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}

	_, body, _ = testRequest(t, testGetBool(manager1), cookie1)
	if body != "false" {
		t.Fatalf("got %q: expected %q", body, "false")
	}
}
