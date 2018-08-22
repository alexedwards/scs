package scs

import (
	"time"
)

// Deprecated: Please use the Manager.Name() method to change the name of the
// session cookie.
var CookieName = "session"

type options struct {
	domain      string
	httpOnly    bool
	idleTimeout time.Duration
	lifetime    time.Duration
	name        string
	path        string
	persist     bool
	secure      bool
	sameSite    string
}
