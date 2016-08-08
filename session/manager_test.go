package session

import (
	"strings"
	"testing"

	"github.com/alexedwards/scs/mem/engine"
)

func TestWriteResponse(t *testing.T) {
	m := Manage(engine.New())
	h := m(testServeMux)

	code, _, _ := testRequest(t, h, "/WriteHeader", "")
	if code != 418 {
		t.Fatalf("got %d: expected %d", code, 418)
	}
}

func TestManagerOptionsLeak(t *testing.T) {
	_ = Manage(engine.New(), Domain("example.org"))

	m := Manage(engine.New())
	h := m(testServeMux)
	_, _, cookie := testRequest(t, h, "/PutString", "")
	if strings.Contains(cookie, "example.org") == true {
		t.Fatalf("got %q: expected to not contain %q", cookie, "example.org")
	}
}
