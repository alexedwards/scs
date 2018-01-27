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
}

type Options interface {
	Domain() string
	HttpOnly() bool
	IdleTimeout() time.Duration
	Lifetime() time.Duration
	Name() string
	Path() string
	Persist() bool
	Secure() bool
}

func (o *options) Domain() string {
	return o.domain
}

func (o *options) HttpOnly() bool {
	return o.httpOnly
}

func (o *options) IdleTimeout() time.Duration {
	return o.idleTimeout
}

func (o *options) Lifetime() time.Duration {
	return o.lifetime
}

func (o *options) Name() string {
	return o.name
}

func (o *options) Path() string {
	return o.path
}

func (o *options) Persist() bool {
	return o.persist
}

func (o *options) Secure() bool {
	return o.secure
}
