package scs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) execute(t *testing.T, urlPath string) (http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.Header, string(body)
}

func extractTokenFromCookie(c string) string {
	parts := strings.Split(c, ";")
	return strings.SplitN(parts[0], "=", 2)[1]
}

func TestEnable(t *testing.T) {
	t.Parallel()

	sessionManager := New()

	mux := http.NewServeMux()
	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := sessionManager.Get(r.Context(), "foo").(string)
		w.Write([]byte(s))
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	header, _ := ts.execute(t, "/put")
	token1 := extractTokenFromCookie(header.Get("Set-Cookie"))

	header, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
	if header.Get("Set-Cookie") != "" {
		t.Errorf("want %q; got %q", "", header.Get("Set-Cookie"))
	}

	header, _ = ts.execute(t, "/put")
	token2 := extractTokenFromCookie(header.Get("Set-Cookie"))
	if token1 != token2 {
		t.Error("want tokens to be the same")
	}
}

func TestLifetime(t *testing.T) {
	t.Parallel()

	sessionManager := New()
	sessionManager.Lifetime = 500 * time.Millisecond

	mux := http.NewServeMux()
	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := sessionManager.Get(r.Context(), "foo")
		if v == nil {
			http.Error(w, "foo does not exist in session", 500)
			return
		}
		w.Write([]byte(v.(string)))
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	ts.execute(t, "/put")

	_, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
	time.Sleep(time.Second)

	_, body = ts.execute(t, "/get")
	if body != "foo does not exist in session\n" {
		t.Errorf("want %q; got %q", "foo does not exist in session\n", body)
	}
}

func TestIdleTimeout(t *testing.T) {
	t.Parallel()

	sessionManager := New()
	sessionManager.IdleTimeout = 200 * time.Millisecond
	sessionManager.Lifetime = time.Second

	mux := http.NewServeMux()
	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := sessionManager.Get(r.Context(), "foo")
		if v == nil {
			http.Error(w, "foo does not exist in session", 500)
			return
		}
		w.Write([]byte(v.(string)))
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	ts.execute(t, "/put")

	time.Sleep(100 * time.Millisecond)
	ts.execute(t, "/get")

	time.Sleep(150 * time.Millisecond)
	_, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}

	time.Sleep(200 * time.Millisecond)
	_, body = ts.execute(t, "/get")
	if body != "foo does not exist in session\n" {
		t.Errorf("want %q; got %q", "foo does not exist in session\n", body)
	}
}

func TestDestroy(t *testing.T) {
	t.Parallel()

	sessionManager := New()

	mux := http.NewServeMux()
	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/destroy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sessionManager.Destroy(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}))
	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := sessionManager.Get(r.Context(), "foo")
		if v == nil {
			http.Error(w, "foo does not exist in session", 500)
			return
		}
		w.Write([]byte(v.(string)))
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	ts.execute(t, "/put")
	header, _ := ts.execute(t, "/destroy")
	cookie := header.Get("Set-Cookie")

	if strings.HasPrefix(cookie, fmt.Sprintf("%s=;", sessionManager.Cookie.Name)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=;", sessionManager.Cookie.Name))
	}
	if strings.Contains(cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}
	if strings.Contains(cookie, "Max-Age=0") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Max-Age=0")
	}

	_, body := ts.execute(t, "/get")
	if body != "foo does not exist in session\n" {
		t.Errorf("want %q; got %q", "foo does not exist in session\n", body)
	}
}

func TestRenewToken(t *testing.T) {
	t.Parallel()

	sessionManager := New()

	mux := http.NewServeMux()
	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/renew", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := sessionManager.RenewToken(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}))
	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := sessionManager.Get(r.Context(), "foo")
		if v == nil {
			http.Error(w, "foo does not exist in session", 500)
			return
		}
		w.Write([]byte(v.(string)))
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	header, _ := ts.execute(t, "/put")
	cookie := header.Get("Set-Cookie")
	originalToken := extractTokenFromCookie(cookie)

	header, _ = ts.execute(t, "/renew")
	cookie = header.Get("Set-Cookie")
	newToken := extractTokenFromCookie(cookie)

	if newToken == originalToken {
		t.Fatal("token has not changed")
	}

	header, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
}

func TestRememberMe(t *testing.T) {
	t.Parallel()

	sessionManager := New()
	sessionManager.Cookie.Persist = false

	mux := http.NewServeMux()
	mux.HandleFunc("/put-normal", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/put-rememberMe-true", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.RememberMe(r.Context(), true)
		sessionManager.Put(r.Context(), "foo", "bar")
	}))
	mux.HandleFunc("/put-rememberMe-false", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionManager.RememberMe(r.Context(), false)
		sessionManager.Put(r.Context(), "foo", "bar")
	}))

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	defer ts.Close()

	header, _ := ts.execute(t, "/put-normal")
	header.Get("Set-Cookie")

	if strings.Contains(header.Get("Set-Cookie"), "Max-Age=") || strings.Contains(header.Get("Set-Cookie"), "Expires=") {
		t.Errorf("want no Max-Age or Expires attributes; got %q", header.Get("Set-Cookie"))
	}

	header, _ = ts.execute(t, "/put-rememberMe-true")
	header.Get("Set-Cookie")

	if !strings.Contains(header.Get("Set-Cookie"), "Max-Age=") || !strings.Contains(header.Get("Set-Cookie"), "Expires=") {
		t.Errorf("want Max-Age and Expires attributes; got %q", header.Get("Set-Cookie"))
	}

	header, _ = ts.execute(t, "/put-rememberMe-false")
	header.Get("Set-Cookie")

	if strings.Contains(header.Get("Set-Cookie"), "Max-Age=") || strings.Contains(header.Get("Set-Cookie"), "Expires=") {
		t.Errorf("want no Max-Age or Expires attributes; got %q", header.Get("Set-Cookie"))
	}
}
