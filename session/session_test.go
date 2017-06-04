package session

import (
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/databrary/scs/engine/memstore"
)

var testEngine Engine
var testServeMux *http.ServeMux

type testUser struct {
	Name string
	Age  int
}

func init() {
	gob.Register(new(testUser))

	testEngine = memstore.New(time.Minute)
	testServeMux = http.NewServeMux()

	testServeMux.HandleFunc("/PutString", func(w http.ResponseWriter, r *http.Request) {
		err := PutString(r, "test_string", "lorem ipsum")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/PersistTrue", func(w http.ResponseWriter, r *http.Request) {
		err := SetPersist(r, true)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		err = PutString(r, "test_string", "lorem ipsum")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/PersistFalse", func(w http.ResponseWriter, r *http.Request) {
		err := SetPersist(r, false)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		err = PutString(r, "test_string", "lorem ipsum")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetString", func(w http.ResponseWriter, r *http.Request) {
		s, err := GetString(r, "test_string")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, s)
	})

	testServeMux.HandleFunc("/PopString", func(w http.ResponseWriter, r *http.Request) {
		s, err := PopString(r, "test_string")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, s)
	})

	testServeMux.HandleFunc("/PutBool", func(w http.ResponseWriter, r *http.Request) {
		err := PutBool(r, "test_bool", true)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetBool", func(w http.ResponseWriter, r *http.Request) {
		b, err := GetBool(r, "test_bool")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%v", b)
	})

	testServeMux.HandleFunc("/PopBool", func(w http.ResponseWriter, r *http.Request) {
		b, err := PopBool(r, "test_bool")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%v", b)
	})

	testServeMux.HandleFunc("/PutInt", func(w http.ResponseWriter, r *http.Request) {
		err := PutInt(r, "test_int", 12345)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetInt", func(w http.ResponseWriter, r *http.Request) {
		i, err := GetInt(r, "test_int")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%d", i)
	})

	testServeMux.HandleFunc("/PopInt", func(w http.ResponseWriter, r *http.Request) {
		i, err := PopInt(r, "test_int")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%d", i)
	})

	testServeMux.HandleFunc("/PutInt64", func(w http.ResponseWriter, r *http.Request) {
		err := PutInt64(r, "test_int", 9223372036854775807)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetInt64", func(w http.ResponseWriter, r *http.Request) {
		i, err := GetInt64(r, "test_int")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%d", i)
	})

	testServeMux.HandleFunc("/PopInt64", func(w http.ResponseWriter, r *http.Request) {
		i, err := PopInt64(r, "test_int")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%d", i)
	})

	testServeMux.HandleFunc("/PutFloat", func(w http.ResponseWriter, r *http.Request) {
		err := PutFloat(r, "test_float", 12.345)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetFloat", func(w http.ResponseWriter, r *http.Request) {
		f, err := GetFloat(r, "test_float")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%.3f", f)
	})

	testServeMux.HandleFunc("/PopFloat", func(w http.ResponseWriter, r *http.Request) {
		f, err := PopFloat(r, "test_float")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%.3f", f)
	})

	testServeMux.HandleFunc("/PutTime", func(w http.ResponseWriter, r *http.Request) {
		err := PutTime(r, "test_time", time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetTime", func(w http.ResponseWriter, r *http.Request) {
		t, err := GetTime(r, "test_time")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, t.Format(time.RFC822))
	})

	testServeMux.HandleFunc("/PopTime", func(w http.ResponseWriter, r *http.Request) {
		t, err := PopTime(r, "test_time")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, t.Format(time.RFC822))
	})

	testServeMux.HandleFunc("/PutBytes", func(w http.ResponseWriter, r *http.Request) {
		err := PutBytes(r, "test_bytes", []byte("lorem ipsum"))
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetBytes", func(w http.ResponseWriter, r *http.Request) {
		b, err := GetBytes(r, "test_bytes")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%s", b)
	})

	testServeMux.HandleFunc("/PopBytes", func(w http.ResponseWriter, r *http.Request) {
		b, err := PopBytes(r, "test_bytes")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%s", b)
	})

	testServeMux.HandleFunc("/PutObject", func(w http.ResponseWriter, r *http.Request) {
		u := &testUser{"alice", 21}
		err := PutObject(r, "test_object", u)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/GetObject", func(w http.ResponseWriter, r *http.Request) {
		u := new(testUser)
		err := GetObject(r, "test_object", u)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%s: %d", u.Name, u.Age)
	})

	testServeMux.HandleFunc("/PopObject", func(w http.ResponseWriter, r *http.Request) {
		u := new(testUser)
		err := PopObject(r, "test_object", u)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%s: %d", u.Name, u.Age)
	})

	testServeMux.HandleFunc("/Keys", func(w http.ResponseWriter, r *http.Request) {
		keys, err := Keys(r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%v", keys)
	})

	testServeMux.HandleFunc("/Exists", func(w http.ResponseWriter, r *http.Request) {
		exists, err := Exists(r, "test_string")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fmt.Fprintf(w, "%v", exists)
	})

	testServeMux.HandleFunc("/RemoveString", func(w http.ResponseWriter, r *http.Request) {
		err := Remove(r, "test_string")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/Clear", func(w http.ResponseWriter, r *http.Request) {
		err := Clear(r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/Destroy", func(w http.ResponseWriter, r *http.Request) {
		err := Destroy(w, r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/RegenerateToken", func(w http.ResponseWriter, r *http.Request) {
		err := RegenerateToken(r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/Renew", func(w http.ResponseWriter, r *http.Request) {
		err := Renew(r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/Save", func(w http.ResponseWriter, r *http.Request) {
		err := PutString(r, "test_string", "lorem ipsum")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		err = Save(w, r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		io.WriteString(w, "OK")
	})

	testServeMux.HandleFunc("/Flush", func(w http.ResponseWriter, r *http.Request) {
		err := PutString(r, "test_string", "lorem ipsum")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		err = Save(w, r)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		fw, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "could not assert to Flusher", 500)
			return
		}
		w.Write([]byte("This is some…"))
		fw.Flush()
		w.Write([]byte("flushed data"))
	})

	testServeMux.HandleFunc("/WriteHeader", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		io.WriteString(w, http.StatusText(http.StatusTeapot))
	})

	testServeMux.HandleFunc("/DumpContext", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", r.Context())
	})
}

func testRequest(t *testing.T, h http.Handler, path string, cookie string) (int, string, string) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", path, nil)
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
	return strings.TrimPrefix(parts[0], fmt.Sprintf("%s=", CookieName))
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

func TestDestroy(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	oldToken := extractTokenFromCookie(cookie)

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/Destroy", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("Cookie", cookie)
	h.ServeHTTP(rr, r)
	body := string(rr.Body.Bytes())
	cookie = rr.Header().Get("Set-Cookie")
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	if len(rr.Header()["Set-Cookie"]) != 1 {
		t.Fatalf("got %d: expected %d", len(rr.Header()["Set-Cookie"]), 1)
	}
	if strings.HasPrefix(cookie, fmt.Sprintf("%s=;", CookieName)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=;", CookieName))
	}
	if strings.Contains(cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Expires=Thu, 01 Jan 1970 00:00:01 GMT")
	}
	if strings.Contains(cookie, "Max-Age=0") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Max-Age=0")
	}
	_, found, _ := e.Find(oldToken)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestRegenerateToken(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	oldToken := extractTokenFromCookie(cookie)

	_, body, cookie := testRequest(t, h, "/RegenerateToken", cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatal("expected a difference")
	}
	_, found, _ := e.Find(oldToken)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != "lorem ipsum" {
		t.Fatalf("got %q: expected %q", body, "lorem ipsum")
	}
}

func TestRenew(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	_, _, cookie := testRequest(t, h, "/PutString", "")
	oldToken := extractTokenFromCookie(cookie)

	_, body, cookie := testRequest(t, h, "/Renew", cookie)
	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	newToken := extractTokenFromCookie(cookie)
	if newToken == oldToken {
		t.Fatal("expected a difference")
	}
	_, found, _ := e.Find(oldToken)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}

	_, body, _ = testRequest(t, h, "/GetString", cookie)
	if body != "" {
		t.Fatalf("got %q: expected %q", body, "")
	}
}

func TestSave(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/Save", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(rr, r)

	body := string(rr.Body.Bytes())
	cookie := rr.Header().Get("Set-Cookie")
	token := extractTokenFromCookie(cookie)

	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
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

func TestSetPersistTrue(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/PersistTrue", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(rr, r)

	body := string(rr.Body.Bytes())
	cookie := rr.Header().Get("Set-Cookie")
	token := extractTokenFromCookie(cookie)

	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	if len(rr.Header()["Set-Cookie"]) != 1 {
		t.Fatalf("got %d: expected %d", len(rr.Header()["Set-Cookie"]), 1)
	}
	if strings.HasPrefix(cookie, fmt.Sprintf("%s=", CookieName)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=", CookieName))
	}
	b, found, _ := e.Find(token)
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	if strings.Contains(cookie, "Expires") == false {
		t.Fatalf("got %q: expected to contain %q", cookie, "Expires... something")
	}

	_, _, persist, err := decodeFromJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if persist != true {
		t.Fatalf("got %q: expected to contain %q", persist, "true")
	}

}

func TestSetPersistFalse(t *testing.T) {
	e := testEngine
	m := Manage(e)
	h := m(testServeMux)

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/PersistFalse", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(rr, r)

	body := string(rr.Body.Bytes())
	cookie := rr.Header().Get("Set-Cookie")
	token := extractTokenFromCookie(cookie)

	if body != "OK" {
		t.Fatalf("got %q: expected %q", body, "OK")
	}
	if len(rr.Header()["Set-Cookie"]) != 1 {
		t.Fatalf("got %d: expected %d", len(rr.Header()["Set-Cookie"]), 1)
	}
	if strings.HasPrefix(cookie, fmt.Sprintf("%s=", CookieName)) == false {
		t.Fatalf("got %q: expected prefix %q", cookie, fmt.Sprintf("%s=", CookieName))
	}
	b, found, _ := e.Find(token)
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	if strings.Contains(cookie, "Expires") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "Expires")
	}

	_, _, persist, err := decodeFromJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if persist != false {
		t.Fatalf("got %q: expected to contain %q", persist, "false")
	}
}
