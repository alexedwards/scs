package session

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCookieOptions(t *testing.T) {
	m := Manage(testEngine)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
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

	m = Manage(testEngine, Path("/foo"), Domain("example.org"), Secure(true), HttpOnly(false), Lifetime(time.Hour), Persist(true))
	h = m(testServeMux)

	_, _, cookie = testRequest(t, h, "/PutString", "")
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

	m = Manage(testEngine, Lifetime(time.Hour))
	h = m(testServeMux)

	_, _, cookie = testRequest(t, h, "/PutString", "")
	if strings.Contains(cookie, "Max-Age=") == true {
		t.Fatalf("got %q: expected not to contain %q:", cookie, "Max-Age=")
	}
	if strings.Contains(cookie, "Expires=") == true {
		t.Fatalf("got %q: expected not to contain %q:", cookie, "Expires")
	}
}

func TestLifetime(t *testing.T) {
	m := Manage(testEngine, Lifetime(200*time.Millisecond))
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	oldToken := extractTokenFromCookie(cookie)
	time.Sleep(100 * time.Millisecond)

	_, _, cookie = testRequest(t, h, "/PutString", cookie)
	time.Sleep(100 * time.Millisecond)

	_, body, _ := testRequest(t, h, "/GetString", cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
	_, _, cookie = testRequest(t, h, "/PutString", cookie)
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatalf("expected a difference")
	}
}

func TestIdleTimeout(t *testing.T) {
	m := Manage(testEngine, IdleTimeout(100*time.Millisecond), Lifetime(500*time.Millisecond))
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	oldToken := extractTokenFromCookie(cookie)
	time.Sleep(150 * time.Millisecond)

	_, body, _ := testRequest(t, h, "/GetString", cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
	_, _, cookie = testRequest(t, h, "/PutString", cookie)
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatalf("expected a difference")
	}

	_, _, cookie = testRequest(t, h, "/PutString", "")
	oldToken = extractTokenFromCookie(cookie)
	time.Sleep(75 * time.Millisecond)

	_, _, cookie = testRequest(t, h, "/GetString", cookie)
	time.Sleep(75 * time.Millisecond)

	_, body, cookie = testRequest(t, h, "/GetString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}
	_, _, cookie = testRequest(t, h, "/PutString", cookie)
	newToken = extractTokenFromCookie(cookie)
	if newToken != oldToken {
		t.Fatalf("expected the same")
	}
}

func TestErrorFunc(t *testing.T) {
	m := Manage(testEngine)
	man, ok := m(nil).(*manager)
	if ok == false {
		t.Fatal("type assertion failed")
	}

	rr := httptest.NewRecorder()
	man.opts.errorFunc(rr, new(http.Request), errors.New("testing error log..."))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("got %d: expected %d", rr.Code, http.StatusInternalServerError)
	}
	if string(rr.Body.Bytes()) != fmt.Sprintf("%s\n", http.StatusText(http.StatusInternalServerError)) {
		t.Fatalf("got %q: expected %q", string(rr.Body.Bytes()), fmt.Sprintf("%s\n", http.StatusText(http.StatusInternalServerError)))
	}

	m = Manage(testEngine, ErrorFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(418)
		io.WriteString(w, http.StatusText(418))
	}))
	man, ok = m(nil).(*manager)
	if ok == false {
		t.Fatal("type assertion failed")
	}

	rr = httptest.NewRecorder()
	man.opts.errorFunc(rr, new(http.Request), errors.New("testing error log..."))
	if rr.Code != 418 {
		t.Fatalf("got %d: expected %d", rr.Code, 418)
	}
	if string(rr.Body.Bytes()) != http.StatusText(418) {
		t.Fatalf("got %q: expected %q", string(rr.Body.Bytes()), http.StatusText(418))
	}
}

func TestPersist(t *testing.T) {
	m := Manage(testEngine, IdleTimeout(5*time.Minute), Persist(true))
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	if strings.Contains(cookie, "Max-Age=300") == false {
		t.Fatalf("got %q: expected to contain %q:", cookie, "Max-Age=300")
	}
}

func TestCookieName(t *testing.T) {
	oldCookieName := CookieName
	CookieName = "custom_cookie_name"

	m := Manage(testEngine)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	if strings.HasPrefix(cookie, "custom_cookie_name=") == false {
		t.Fatalf("got %q: expected prefix %q", cookie, "custom_cookie_name=")
	}

	_, body, _ := testRequest(t, h, "/GetString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	CookieName = oldCookieName
}

func TestContextDataName(t *testing.T) {
	oldContextName := ContextName
	ContextName = "custom_context_name"

	m := Manage(testEngine)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	_, body, _ := testRequest(t, h, "/DumpContext", cookie)
	if strings.Contains(body, "custom_context_name") == false {
		t.Fatalf("got %q: expected to contain %q", body, "custom_context_name")
	}
	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}

	ContextName = oldContextName
}
