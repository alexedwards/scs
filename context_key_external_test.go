package scs_test

import (
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestContextKeyCanBeSetFromExternalPackage(t *testing.T) {
	sm := scs.New()
	sm.ContextKey = "myapp.session"

	if sm.ContextKey != "myapp.session" {
		t.Fatalf("got context key %q, want %q", sm.ContextKey, "myapp.session")
	}
}
