package scs_test

import (
	"reflect"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestAsInterface(t *testing.T) {
	t.Parallel()
	sessionManager := &scs.SessionManager{}

	castedSession := sessionManager.AsInterface()
	value, ok := castedSession.(*scs.SessionManager)
	if !ok {
		t.Error("unable to cast to SessionManager")
	}

	if !reflect.DeepEqual(sessionManager, value) {
		t.Errorf("want %+v; got %+v", sessionManager, value)
	}
}
