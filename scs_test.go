package scs

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, h http.Handler, cookie string) (int, string, string) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if cookie != "" {
		r.Header.Add("Cookie", cookie)
	}
	h.ServeHTTP(rr, r)

	code := rr.Code
	body := string(rr.Body.Bytes())
	cookie = rr.Header().Get("Set-Cookie")
	return code, body, cookie
}

func extractTokenFromCookie(c string) string {
	parts := strings.Split(c, ";")
	return strings.SplitN(parts[0], "=", 2)[1]
}

// Test Handlers

func testPutString(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		err := session.PutString(w, "test_string", "lorem ipsum")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testGetString(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		s, err := session.GetString("test_string")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, s)
	}
}

func testPopString(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		s, err := session.PopString(w, "test_string")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, s)
	}
}

func testPutBool(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		err := session.PutBool(w, "test_bool", true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testGetBool(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		b, err := session.GetBool("test_bool")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%v", b)
	}
}

func testPutObject(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		u := &testUser{"alice", 21}
		err := session.PutObject(w, "test_object", u)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testGetObject(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		u := new(testUser)
		err := session.GetObject("test_object", u)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%s: %d", u.Name, u.Age)
	}
}

func testPopObject(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		u := new(testUser)
		err := session.PopObject(w, "test_object", u)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%s: %d", u.Name, u.Age)
	}
}

func testDestroy(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		err := session.Destroy(w)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testRenewToken(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		err := session.RenewToken(w)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testClear(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		err := session.Clear(w)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, "OK")
	}
}

func testKeys(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := manager.Load(r)
		keys, err := session.Keys()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "%v", keys)
	}
}
