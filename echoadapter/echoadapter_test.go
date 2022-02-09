package echoadapter

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
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

	sessionManager := scs.New()

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/get", func(c echo.Context) error {
		s := sessionManager.Get(c.Request().Context(), "foo").(string)
		return c.String(http.StatusOK, s)
	})

	ts := newTestServer(t, e)
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

	sessionManager := scs.New()
	sessionManager.Lifetime = 500 * time.Millisecond

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/get", func(c echo.Context) error {
		v := sessionManager.Get(c.Request().Context(), "foo")
		if v == nil {
			return c.String(500, "foo does not exist in session")
		}
		return c.String(http.StatusOK, v.(string))
	})

	ts := newTestServer(t, e)
	defer ts.Close()

	ts.execute(t, "/put")

	_, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
	time.Sleep(time.Second)

	_, body = ts.execute(t, "/get")
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestIdleTimeout(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()
	sessionManager.IdleTimeout = 200 * time.Millisecond
	sessionManager.Lifetime = time.Second

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/get", func(c echo.Context) error {
		v := sessionManager.Get(c.Request().Context(), "foo")
		if v == nil {
			return c.String(500, "foo does not exist in session")
		}
		return c.String(http.StatusOK, v.(string))
	})

	ts := newTestServer(t, e)
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
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestDestroy(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/destroy", func(c echo.Context) error {
		err := sessionManager.Destroy(c.Request().Context())
		if err != nil {
			return c.String(500, err.Error())
		}
		return nil
	})
	e.GET("/get", func(c echo.Context) error {
		v := sessionManager.Get(c.Request().Context(), "foo")
		if v == nil {
			return c.String(500, "foo does not exist in session")
		}
		return c.String(http.StatusOK, v.(string))
	})

	ts := newTestServer(t, e)
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
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestRenewToken(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/renew", func(c echo.Context) error {
		err := sessionManager.RenewToken(c.Request().Context())
		if err != nil {
			return c.String(500, err.Error())
		}
		return nil
	})
	e.GET("/get", func(c echo.Context) error {
		v := sessionManager.Get(c.Request().Context(), "foo")
		if v == nil {
			return c.String(500, "foo does not exist in session")
		}
		return c.String(http.StatusOK, v.(string))
	})

	ts := newTestServer(t, e)
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

	_, body := ts.execute(t, "/get")
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
}

func TestRememberMe(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()
	sessionManager.Cookie.Persist = false

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put-normal", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/put-rememberMe-true", func(c echo.Context) error {
		sessionManager.RememberMe(c.Request().Context(), true)
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})
	e.GET("/put-rememberMe-false", func(c echo.Context) error {
		sessionManager.RememberMe(c.Request().Context(), false)
		sessionManager.Put(c.Request().Context(), "foo", "bar")
		return nil
	})

	ts := newTestServer(t, e)
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

func TestIterate(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	e := echo.New()

	sessionAdapter := New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", func(c echo.Context) error {
		sessionManager.Put(c.Request().Context(), "foo", c.QueryParam("foo"))
		return nil
	})

	for i := 0; i < 3; i++ {
		ts := newTestServer(t, e)
		defer ts.Close()

		ts.execute(t, "/put?foo="+strconv.Itoa(i))
	}

	results := []string{}

	err := sessionManager.Iterate(context.Background(), func(ctx context.Context) error {
		i := sessionManager.GetString(ctx, "foo")
		results = append(results, i)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(results)

	if !reflect.DeepEqual(results, []string{"0", "1", "2"}) {
		t.Fatalf("unexpected value: got %v", results)
	}

	err = sessionManager.Iterate(context.Background(), func(ctx context.Context) error {
		return errors.New("expected error")
	})
	if err.Error() != "expected error" {
		t.Fatal("didn't get expected error")
	}
}
