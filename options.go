package scs

import (
	"time"
)

// CookieName changes the name of the session cookie issued to clients. Note that
// cookie names should not contain whitespace, commas, semicolons, backslashes
// or control characters as per RFC6265.
var CookieName = "session"

type options struct {
	domain      string
	httpOnly    bool
	idleTimeout time.Duration
	lifetime    time.Duration
	path        string
	persist     bool
	secure      bool
}
