package scs

import (
	"strings"
	"testing"
	"time"
)

func TestCookieOptions(t *testing.T) {
	manager := NewManager(newMockStore())

	_, _, cookie := testRequest(t, testPutString(manager), "")
	if strings.Contains(cookie, "Path=/") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Path=/")
	}
	if strings.Contains(cookie, "Domain=") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "Domain=")
	}
	if strings.Contains(cookie, "Secure") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "Secure")
	}
	if strings.Contains(cookie, "HttpOnly") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "HttpOnly")
	}

	manager = NewManager(newMockStore())
	manager.Path("/foo")
	manager.Domain("example.org")
	manager.Secure(true)
	manager.HttpOnly(false)
	manager.Lifetime(time.Hour)
	manager.Persist(true)

	_, _, cookie = testRequest(t, testPutString(manager), "")
	if strings.Contains(cookie, "Path=/foo") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Path=/foo")
	}
	if strings.Contains(cookie, "Domain=example.org") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Domain=example.org")
	}
	if strings.Contains(cookie, "Secure") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Secure")
	}
	if strings.Contains(cookie, "HttpOnly") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "HttpOnly")
	}
	if strings.Contains(cookie, "Max-Age=3600") == false {
		t.Fatalf("got %q: expected to contain %q:", cookie, "Max-Age=86400")
	}
	if strings.Contains(cookie, "Expires=") == false {
		t.Fatalf("got %q: expected to contain %q:", cookie, "Expires")
	}

	manager = NewManager(newMockStore())
	manager.Lifetime(time.Hour)

	_, _, cookie = testRequest(t, testPutString(manager), "")
	if strings.Contains(cookie, "Max-Age=") == true {
		t.Fatalf("got %q: expected not to contain %q:", cookie, "Max-Age=")
	}
	if strings.Contains(cookie, "Expires=") == true {
		t.Fatalf("got %q: expected not to contain %q:", cookie, "Expires")
	}
}

func TestLifetime(t *testing.T) {
	manager := NewManager(newMockStore())
	manager.Lifetime(200 * time.Millisecond)

	_, _, cookie := testRequest(t, testPutString(manager), "")
	oldToken := extractTokenFromCookie(cookie)
	time.Sleep(100 * time.Millisecond)

	_, _, cookie = testRequest(t, testPutString(manager), cookie)
	time.Sleep(100 * time.Millisecond)

	_, body, _ := testRequest(t, testGetString(manager), cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
	_, _, cookie = testRequest(t, testPutString(manager), cookie)
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatalf("expected a difference")
	}
}

func TestIdleTimeout(t *testing.T) {
	manager := NewManager(newMockStore())
	manager.IdleTimeout(100 * time.Millisecond)
	manager.Lifetime(500 * time.Millisecond)

	_, _, cookie := testRequest(t, testPutString(manager), "")
	oldToken := extractTokenFromCookie(cookie)
	time.Sleep(150 * time.Millisecond)

	_, body, _ := testRequest(t, testGetString(manager), cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
	_, _, cookie = testRequest(t, testPutString(manager), cookie)
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatalf("expected a difference")
	}

	_, _, cookie = testRequest(t, testPutString(manager), "")
	oldToken = extractTokenFromCookie(cookie)
	time.Sleep(75 * time.Millisecond)

	_, _, cookie = testRequest(t, testPutString(manager), cookie)
	time.Sleep(75 * time.Millisecond)

	_, body, _ = testRequest(t, testGetString(manager), cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}
	_, _, cookie = testRequest(t, testPutString(manager), cookie)
	newToken = extractTokenFromCookie(cookie)
	if newToken != oldToken {
		t.Fatalf("expected the same")
	}
}

func TestPersist(t *testing.T) {
	manager := NewManager(newMockStore())
	manager.IdleTimeout(5 * time.Minute)
	manager.Persist(true)

	_, _, cookie := testRequest(t, testPutString(manager), "")
	if strings.Contains(cookie, "Max-Age=300") == false {
		t.Fatalf("got %q: expected to contain %q:", cookie, "Max-Age=300")
	}
}

func TestName(t *testing.T) {
	manager := NewManager(newMockStore())
	manager.Name("foo")

	_, _, cookie := testRequest(t, testPutString(manager), "")
	if strings.HasPrefix(cookie, "foo=") == false {
		t.Fatalf("got %q: expected prefix %q", cookie, "foo=")
	}

	_, body, _ := testRequest(t, testGetString(manager), cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}
}
