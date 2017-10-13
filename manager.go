package scs

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/stores/cookiestore"
)

// Manager is a session manager.
type Manager struct {
	store Store
	opts  *options
}

// NewManager returns a pointer to a new session manager.
func NewManager(store Store) *Manager {
	defaultOptions := &options{
		domain:      "",
		httpOnly:    true,
		idleTimeout: 0,
		lifetime:    24 * time.Hour,
		name:        "session",
		path:        "/",
		persist:     false,
		secure:      false,
	}

	return &Manager{
		store: store,
		opts:  defaultOptions,
	}
}

// Domain sets the 'Domain' attribute on the session cookie. By default it will
// be set to the domain name that the cookie was issued from.
func (m *Manager) Domain(s string) {
	m.opts.domain = s
}

// HttpOnly sets the 'HttpOnly' attribute on the session cookie. The default value
// is true.
func (m *Manager) HttpOnly(b bool) {
	m.opts.httpOnly = b
}

// IdleTimeout sets the maximum length of time a session can be inactive before it
// expires. For example, some applications may wish to set this so there is a timeout
// after 20 minutes of inactivity. The inactivity period is reset whenever the
// session data is changed (but not read).
//
// By default IdleTimeout is not set and there is no inactivity timeout.
func (m *Manager) IdleTimeout(t time.Duration) {
	m.opts.idleTimeout = t
}

// Lifetime sets the maximum length of time that a session is valid for before
// it expires. The lifetime is an 'absolute expiry' which is set when the session
// is first created and does not change.
//
// The default value is 24 hours.
func (m *Manager) Lifetime(t time.Duration) {
	m.opts.lifetime = t
}

// Name sets the name of the session cookie. This name should not contain whitespace,
// commas, semicolons, backslashes, the equals sign or control characters as per
// RFC6265.
func (m *Manager) Name(s string) {
	m.opts.name = s
}

// Path sets the 'Path' attribute on the session cookie. The default value is "/".
// Passing the empty string "" will result in it being set to the path that the
// cookie was issued from.
func (m *Manager) Path(s string) {
	m.opts.path = s
}

// Persist sets whether the session cookie should be persistent or not (i.e. whether
// it should be retained after a user closes their browser).
//
// The default value is false, which means that the session cookie will be destroyed
// when the user closes their browser. If set to true, explicit 'Expires' and
// 'MaxAge' values will be added to the cookie and it will be retained by the
// user's browser until the given expiry time is reached.
func (m *Manager) Persist(b bool) {
	m.opts.persist = b
}

// Secure sets the 'Secure' attribute on the session cookie. The default value
// is false. It's recommended that you set this to true and serve all requests
// over HTTPS in production environments.
func (m *Manager) Secure(b bool) {
	m.opts.secure = b
}

// Load returns the session data for the current request.
func (m *Manager) Load(r *http.Request) *Session {
	return load(r, m.store, m.opts)
}

func NewCookieManager(key string) *Manager {
	store := cookiestore.New([]byte(key))
	return NewManager(store)
}

func (m *Manager) Multi(next http.Handler) http.Handler {
	return m.Use(next)
}

func (m *Manager) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := m.Load(r)

		if m.opts.idleTimeout > 0 {
			err := session.Touch(w)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		ctx := context.WithValue(r.Context(), sessionName(m.opts.name), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
