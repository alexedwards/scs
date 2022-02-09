package fiberadapter

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gofiber/fiber/v2"
)

type testServer struct {
	*fiber.App
}

func newTestServer(t *testing.T, app *fiber.App) *testServer {
	return &testServer{app}
}

func (ts *testServer) execute(t *testing.T, urlPath string, reqCookie *http.Cookie) (*http.Cookie, string) {
	req := httptest.NewRequest("GET", urlPath, nil)
	if reqCookie != nil {
		req.AddCookie(reqCookie)
	}

	res, err := ts.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var resCookie *http.Cookie
	if len(res.Cookies()) > 0 {
		resCookie = res.Cookies()[0]
	}

	return resCookie, string(body)
}

func TestEnable(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		s := sessionManager.Get(c.UserContext(), "foo").(string)
		return c.SendString(s)
	})

	ts := newTestServer(t, app)

	cookie1, _ := ts.execute(t, "/put", nil)
	token1 := cookie1.Value

	cookie2, body := ts.execute(t, "/get", cookie1)
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
	if cookie2.String() != "" {
		t.Errorf("want %q; got %q", "", cookie2.String())
	}

	cookie3, _ := ts.execute(t, "/put", cookie1)
	token2 := cookie3.Value
	if token1 != token2 {
		t.Error("want tokens to be the same")
	}
}

func TestLifetime(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()
	sessionManager.Lifetime = 500 * time.Millisecond

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		v := sessionManager.Get(c.UserContext(), "foo")
		if v == nil {
			return c.SendString("foo does not exist in session")
		}
		return c.SendString(v.(string))
	})

	ts := newTestServer(t, app)

	cookie, _ := ts.execute(t, "/put", nil)

	_, body := ts.execute(t, "/get", cookie)
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
	time.Sleep(time.Second)

	_, body = ts.execute(t, "/get", cookie)
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestIdleTimeout(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()
	sessionManager.IdleTimeout = 200 * time.Millisecond
	sessionManager.Lifetime = time.Second

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		v := sessionManager.Get(c.UserContext(), "foo")
		if v == nil {
			return c.SendString("foo does not exist in session")
		}
		return c.SendString(v.(string))
	})

	ts := newTestServer(t, app)

	cookie, _ := ts.execute(t, "/put", nil)

	time.Sleep(100 * time.Millisecond)
	ts.execute(t, "/get", cookie)

	time.Sleep(150 * time.Millisecond)
	_, body := ts.execute(t, "/get", cookie)
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}

	time.Sleep(200 * time.Millisecond)
	_, body = ts.execute(t, "/get", cookie)
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestDestroy(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/destroy", func(c *fiber.Ctx) error {
		err := sessionManager.Destroy(c.UserContext())
		if err != nil {
			return c.SendStatus(500)
		}
		return nil
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		v := sessionManager.Get(c.UserContext(), "foo")
		if v == nil {
			return c.SendString("foo does not exist in session")
		}
		return c.SendString(v.(string))
	})

	ts := newTestServer(t, app)

	cookie, _ := ts.execute(t, "/put", nil)
	cookie, _ = ts.execute(t, "/destroy", cookie)

	if strings.HasPrefix(cookie.String(), fmt.Sprintf("%s=;", sessionManager.Cookie.Name)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=;", sessionManager.Cookie.Name))
	}
	if strings.Contains(cookie.String(), "Expires=Thu, 01 Jan 1970 00:00:01 GMT") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}
	if strings.Contains(cookie.String(), "Max-Age=0") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Max-Age=0")
	}

	_, body := ts.execute(t, "/get", cookie)
	if body != "foo does not exist in session" {
		t.Errorf("want %q; got %q", "foo does not exist in session", body)
	}
}

func TestRenewToken(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/renew", func(c *fiber.Ctx) error {
		err := sessionManager.RenewToken(c.UserContext())
		if err != nil {
			return c.SendStatus(500)
		}
		return nil
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		v := sessionManager.Get(c.UserContext(), "foo")
		if v == nil {
			return c.SendString("foo does not exist in session")
		}
		return c.SendString(v.(string))
	})

	ts := newTestServer(t, app)

	cookie, _ := ts.execute(t, "/put", nil)
	originalToken := cookie.Value

	cookie, _ = ts.execute(t, "/renew", cookie)
	newToken := cookie.Value

	if newToken == originalToken {
		t.Fatal("token has not changed")
	}

	_, body := ts.execute(t, "/get", cookie)
	if body != "bar" {
		t.Errorf("want %q; got %q", "bar", body)
	}
}

func TestRememberMe(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()
	sessionManager.Cookie.Persist = false

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put-normal", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/put-rememberMe-true", func(c *fiber.Ctx) error {
		sessionManager.RememberMe(c.UserContext(), true)
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})
	app.Get("/put-rememberMe-false", func(c *fiber.Ctx) error {
		sessionManager.RememberMe(c.UserContext(), false)
		sessionManager.Put(c.UserContext(), "foo", "bar")
		return nil
	})

	ts := newTestServer(t, app)

	cookie, _ := ts.execute(t, "/put-normal", nil)

	if strings.Contains(cookie.String(), "Max-Age=") || strings.Contains(cookie.String(), "Expires=") {
		t.Errorf("want no Max-Age or Expires attributes; got %q", cookie.String())
	}

	cookie, _ = ts.execute(t, "/put-rememberMe-true", cookie)

	if !strings.Contains(cookie.String(), "Max-Age=") || !strings.Contains(cookie.String(), "Expires=") {
		t.Errorf("want Max-Age and Expires attributes; got %q", cookie.String())
	}

	cookie, _ = ts.execute(t, "/put-rememberMe-false", cookie)

	if strings.Contains(cookie.String(), "Max-Age=") || strings.Contains(cookie.String(), "Expires=") {
		t.Errorf("want no Max-Age or Expires attributes; got %q", cookie.String())
	}
}

func TestIterate(t *testing.T) {
	t.Parallel()

	sessionManager := scs.New()

	app := fiber.New()

	sessionAdapter := New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", func(c *fiber.Ctx) error {
		sessionManager.Put(c.UserContext(), "foo", c.Query("foo"))
		return nil
	})

	for i := 0; i < 3; i++ {
		ts := newTestServer(t, app)

		ts.execute(t, "/put?foo="+strconv.Itoa(i), nil)
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
