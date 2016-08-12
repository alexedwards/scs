package session

import (
	"net/http"
	"time"
)

var (
	ContextDataName = "session.data"
	CookieName      = "session.cookie"
	defaultOptions  = &options{
		domain:    "",
		errorFunc: defaultErrorFunc,
		httpOnly:  true,
		lifetime:  24 * time.Hour,
		path:      "/",
		persist:   false,
		secure:    false,
	}
)

type options struct {
	domain    string
	errorFunc func(http.ResponseWriter, *http.Request, error)
	httpOnly  bool
	lifetime  time.Duration
	path      string
	persist   bool
	secure    bool
}

type Option func(*options)

func Domain(s string) Option {
	return func(opts *options) {
		opts.domain = s
	}
}

func ErrorFunc(f func(http.ResponseWriter, *http.Request, error)) Option {
	return func(opts *options) {
		opts.errorFunc = f
	}
}

func HttpOnly(b bool) Option {
	return func(opts *options) {
		opts.httpOnly = b
	}
}

func Lifetime(t time.Duration) Option {
	return func(opts *options) {
		opts.lifetime = t
	}
}

func Path(s string) Option {
	return func(opts *options) {
		opts.path = s
	}
}

func Persist(b bool) Option {
	return func(opts *options) {
		opts.persist = b
	}
}

func Secure(b bool) Option {
	return func(opts *options) {
		opts.secure = b
	}
}
