package session

import (
	"net/http"
	"time"
)

// ContextName changes the value of the (string) key used to store the session
// information in Request.Context. You should only need to change this if there is
// a naming clash.
var ContextName = "scs.session"

// CookieName changes the name of the session cookie issued to clients. Note that
// cookie names should not contain whitespace, commas, semicolons, backslashes
// or control characters as per RFC6265.
var CookieName = "scs.session.token"

var defaultOptions = &options{
	domain:      "",
	errorFunc:   defaultErrorFunc,
	httpOnly:    true,
	idleTimeout: 0,
	lifetime:    24 * time.Hour,
	path:        "/",
	persist:     false,
	secure:      false,
}

type options struct {
	domain      string
	errorFunc   func(http.ResponseWriter, *http.Request, error)
	httpOnly    bool
	idleTimeout time.Duration
	lifetime    time.Duration
	path        string
	persist     bool
	secure      bool
}

// Option defines the functional arguments for configuring the session manager.
type Option func(*options)

// Domain sets the 'Domain' attribute on the session cookie. By default it will
// be set to the domain name that the cookie was issued from.
func Domain(s string) Option {
	return func(opts *options) {
		opts.domain = s
	}
}

// ErrorFunc allows you to control behavior when an error is encountered loading
// or writing a session. The default behavior is for a HTTP 500 status code to
// be written to the ResponseWriter along with the plain-text error string. If
// a custom error function is set, then control will be passed to this instead.
// A typical use would be to provide a function which logs the error and returns
// a customized HTML error page.
func ErrorFunc(f func(http.ResponseWriter, *http.Request, error)) Option {
	return func(opts *options) {
		opts.errorFunc = f
	}
}

// HttpOnly sets the 'HttpOnly' attribute on the session cookie. The default value
// is true.
func HttpOnly(b bool) Option {
	return func(opts *options) {
		opts.httpOnly = b
	}
}

// IdleTimeout sets the maximum length of time a session can be inactive before it
// expires. For example, some applications may wish to set this so there is a timeout after
// 20 minutes of inactivity. Any client request which includes the
// session cookie and is handled by the session middleware is classed as activity.s
//
// By default IdleTimeout is not set and there is no inactivity timeout.
func IdleTimeout(t time.Duration) Option {
	return func(opts *options) {
		opts.idleTimeout = t
	}
}

// Lifetime sets the maximum length of time that a session is valid for before
// it expires. The lifetime is an 'absolute expiry' which is set when the session
// is first created and does not change.
//
// The default value is 24 hours.
func Lifetime(t time.Duration) Option {
	return func(opts *options) {
		opts.lifetime = t
	}
}

// Path sets the 'Path' attribute on the session cookie. The default value is "/".
// Passing the empty string "" will result in it being set to the path that the
// cookie was issued from.
func Path(s string) Option {
	return func(opts *options) {
		opts.path = s
	}
}

// Persist sets whether the session cookie should be persistent or not (i.e. whether
// it should be retained after a user closes their browser).
//
// The default value is false, which means that the session cookie will be destroyed
// when the user closes their browser. If set to true, explicit 'Expires' and
// 'MaxAge' values will be added to the cookie and it will be retained by the
// user's browser until the given expiry time is reached.
func Persist(b bool) Option {
	return func(opts *options) {
		opts.persist = b
	}
}

// Secure sets the 'Secure' attribute on the session cookie. The default value
// is false. It's recommended that you set this to true and serve all requests
// over HTTPS in production environments.
func Secure(b bool) Option {
	return func(opts *options) {
		opts.secure = b
	}
}
