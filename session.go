package scs

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2/memstore"
)

// Deprecated: Session is a backwards-compatible alias for SessionManager.
type Session = SessionManager

// SessionManager holds the configuration settings for your sessions.
type SessionManager struct {
	// IdleTimeout controls the maximum length of time a session can be inactive
	// before it expires. For example, some applications may wish to set this so
	// there is a timeout after 20 minutes of inactivity. By default IdleTimeout
	// is not set and there is no inactivity timeout.
	IdleTimeout time.Duration

	// Lifetime controls the maximum length of time that a session is valid for
	// before it expires. The lifetime is an 'absolute expiry' which is set when
	// the session is first created and does not change. The default value is 24
	// hours.
	Lifetime time.Duration

	// Store controls the session store where the session data is persisted.
	Store Store

	// Cookie contains the configuration settings for session cookies.
	Cookie SessionCookie

	// Codec controls the encoder/decoder used to transform session data to a
	// byte slice for use by the session store. By default session data is
	// encoded/decoded using encoding/gob.
	Codec Codec

	// ErrorFunc allows you to control behavior when an error is encountered by
	// the LoadAndSave middleware. The default behavior is for a HTTP 500
	// "Internal Server Error" message to be sent to the client and the error
	// logged using Go's standard logger. If a custom ErrorFunc is set, then
	// control will be passed to this instead. A typical use would be to provide
	// a function which logs the error and returns a customized HTML error page.
	ErrorFunc func(http.ResponseWriter, *http.Request, error)

	// HashTokenInStore controls whether or not to store the session token or a hashed version in the store.
	HashTokenInStore bool

	// contextKey is the key used to set and retrieve the session data from a
	// context.Context. It's automatically generated to ensure uniqueness.
	contextKey contextKey
}

// SessionCookie contains the configuration settings for session cookies.
type SessionCookie struct {
	// Name sets the name of the session cookie. It should not contain
	// whitespace, commas, colons, semicolons, backslashes, the equals sign or
	// control characters as per RFC6265. The default cookie name is "session".
	// If your application uses two different sessions, you must make sure that
	// the cookie name for each is unique.
	Name string

	// Domain sets the 'Domain' attribute on the session cookie. By default
	// it will be set to the domain name that the cookie was issued from.
	Domain string

	// HttpOnly sets the 'HttpOnly' attribute on the session cookie. The
	// default value is true.
	HttpOnly bool

	// Path sets the 'Path' attribute on the session cookie. The default value
	// is "/". Passing the empty string "" will result in it being set to the
	// path that the cookie was issued from.
	Path string

	// SameSite controls the value of the 'SameSite' attribute on the session
	// cookie. By default this is set to 'SameSite=Lax'. If you want no SameSite
	// attribute or value in the session cookie then you should set this to 0.
	SameSite http.SameSite

	// Secure sets the 'Secure' attribute on the session cookie. The default
	// value is false. It's recommended that you set this to true and serve all
	// requests over HTTPS in production environments.
	// See https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md#transport-layer-security.
	Secure bool

	// Secure sets the 'Partitioned' attribute on the session cookie. The
	// default value is false.
	Partitioned bool

	// Persist sets whether the session cookie should be persistent or not
	// (i.e. whether it should be retained after a user closes their browser).
	// The default value is true, which means that the session cookie will not
	// be destroyed when the user closes their browser and the appropriate
	// 'Expires' and 'MaxAge' values will be added to the session cookie. If you
	// want to only persist some sessions (rather than all of them), then set this
	// to false and call the RememberMe() method for the specific sessions that you
	// want to persist.
	Persist bool
}

// New returns a new session manager with the default options. It is safe for
// concurrent use.
func New() *SessionManager {
	s := &SessionManager{
		IdleTimeout: 0,
		Lifetime:    24 * time.Hour,
		Store:       memstore.New(),
		Codec:       GobCodec{},
		ErrorFunc:   defaultErrorFunc,
		contextKey:  generateContextKey(),
		Cookie: SessionCookie{
			Name:        "session",
			Domain:      "",
			HttpOnly:    true,
			Path:        "/",
			SameSite:    http.SameSiteLaxMode,
			Secure:      false,
			Partitioned: false,
			Persist:     true,
		},
	}
	return s
}

// Deprecated: NewSession is a backwards-compatible alias for New. Use the New
// function instead.
func NewSession() *SessionManager {
	return New()
}

// LoadAndSave provides middleware which automatically loads and saves session
// data for the current request, and communicates the session token to and from
// the client in a cookie.
func (s *SessionManager) LoadAndSave(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")

		var token string
		cookie, err := r.Cookie(s.Cookie.Name)
		if err == nil {
			token = cookie.Value
		}

		ctx, err := s.Load(r.Context(), token)
		if err != nil {
			s.ErrorFunc(w, r, err)
			return
		}

		sr := r.WithContext(ctx)

		sw := &sessionResponseWriter{
			ResponseWriter: w,
			request:        sr,
			sessionManager: s,
		}

		next.ServeHTTP(sw, sr)

		if !sw.written {
			s.commitAndWriteSessionCookie(w, sr)
		}
	})
}

func (s *SessionManager) commitAndWriteSessionCookie(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch s.Status(ctx) {
	case Modified:
		token, expiry, err := s.Commit(ctx)
		if err != nil {
			s.ErrorFunc(w, r, err)
			return
		}

		s.WriteSessionCookie(ctx, w, token, expiry)
	case Destroyed:
		s.WriteSessionCookie(ctx, w, "", time.Time{})
	}
}

// WriteSessionCookie writes a cookie to the HTTP response with the provided
// token as the cookie value and expiry as the cookie expiry time. The expiry
// time will be included in the cookie only if the session is set to persist
// or has had RememberMe(true) called on it. If expiry is an empty time.Time
// struct (so that it's IsZero() method returns true) the cookie will be
// marked with a historical expiry time and negative max-age (so the browser
// deletes it).
//
// Most applications will use the LoadAndSave() middleware and will not need to
// use this method.
func (s *SessionManager) WriteSessionCookie(ctx context.Context, w http.ResponseWriter, token string, expiry time.Time) {
	cookie := &http.Cookie{
		Value:       token,
		Name:        s.Cookie.Name,
		Domain:      s.Cookie.Domain,
		HttpOnly:    s.Cookie.HttpOnly,
		Path:        s.Cookie.Path,
		SameSite:    s.Cookie.SameSite,
		Secure:      s.Cookie.Secure,
		Partitioned: s.Cookie.Partitioned,
	}

	if expiry.IsZero() {
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else if s.Cookie.Persist || s.GetBool(ctx, "__rememberMe") {
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)        // Round up to the nearest second.
		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1) // Round up to the nearest second.
	}

	w.Header().Add("Set-Cookie", cookie.String())
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}

func defaultErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	log.Output(2, err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type sessionResponseWriter struct {
	http.ResponseWriter
	request        *http.Request
	sessionManager *SessionManager
	written        bool
}

func (sw *sessionResponseWriter) Write(b []byte) (int, error) {
	if !sw.written {
		sw.sessionManager.commitAndWriteSessionCookie(sw.ResponseWriter, sw.request)
		sw.written = true
	}

	return sw.ResponseWriter.Write(b)
}

func (sw *sessionResponseWriter) WriteHeader(code int) {
	if !sw.written {
		sw.sessionManager.commitAndWriteSessionCookie(sw.ResponseWriter, sw.request)
		sw.written = true
	}

	sw.ResponseWriter.WriteHeader(code)
}

func (sw *sessionResponseWriter) Unwrap() http.ResponseWriter {
	return sw.ResponseWriter
}
