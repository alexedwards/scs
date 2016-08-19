package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/engine/memstore"
)

func TestWriteResponse(t *testing.T) {
	m := Manage(memstore.New(time.Minute))
	h := m(testServeMux)

	code, _, _ := testRequest(t, h, "/WriteHeader", "")
	if code != 418 {
		t.Fatalf("got %d: expected %d", code, 418)
	}
}

func TestManagerOptionsLeak(t *testing.T) {
	_ = Manage(memstore.New(time.Minute), Domain("example.org"))

	m := Manage(memstore.New(time.Minute))
	h := m(testServeMux)
	_, _, cookie := testRequest(t, h, "/PutString", "")
	if strings.Contains(cookie, "example.org") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "example.org")
	}
}

func TestFlusher(t *testing.T) {
	e := memstore.New(time.Minute)
	m := Manage(e)
	h := m(testServeMux)

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/Flush", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(rr, r)

	body := string(rr.Body.Bytes())
	cookie := rr.Header().Get("Set-Cookie")
	token := extractTokenFromCookie(cookie)

	if body != "This is some…flushed data" {
		t.Fatalf("got %q: expected %q", body, "This is some…flushed data")
	}
	if len(rr.Header()["Set-Cookie"]) != 1 {
		t.Fatalf("got %d: expected %d", len(rr.Header()["Set-Cookie"]), 1)
	}
	if strings.HasPrefix(cookie, fmt.Sprintf("%s=", CookieName)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=", CookieName))
	}
	_, found, _ := e.Find(token)
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
}
