package session

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/alexedwards/scs"
)

// ErrAlreadyWritten is returned by operations that attempt to modify the
// session data after it has already been sent to the storage engine and client.
var ErrAlreadyWritten = errors.New("session already written to the engine and http.ResponseWriter")

type session struct {
	token    string
	data     map[string]interface{}
	deadline time.Time
	engine   scs.Engine
	opts     *options
	modified bool
	written  bool
	mu       sync.Mutex
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func newSession(r *http.Request, engine scs.Engine, opts *options) (*http.Request, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	s := &session{
		token:    token,
		data:     make(map[string]interface{}),
		deadline: time.Now().Add(opts.lifetime),
		engine:   engine,
		opts:     opts,
	}
	return requestWithSession(r, s), nil
}

func load(r *http.Request, engine scs.Engine, opts *options) (*http.Request, error) {
	cookie, err := r.Cookie(CookieName)
	if err == http.ErrNoCookie {
		return newSession(r, engine, opts)
	} else if err != nil {
		return nil, err
	}

	if cookie.Value == "" {
		return newSession(r, engine, opts)
	}
	token := cookie.Value

	j, found, err := engine.Find(token)
	if err != nil {
		return nil, err
	}
	if found == false {
		return newSession(r, engine, opts)
	}

	data, deadline, err := decodeFromJSON(j)
	if err != nil {
		return nil, err
	}

	s := &session{
		token:    token,
		data:     data,
		deadline: deadline,
		engine:   engine,
		opts:     opts,
	}

	return requestWithSession(r, s), nil
}

func write(w http.ResponseWriter, r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return nil
	}

	if s.modified == false && s.opts.idleTimeout == 0 {
		return nil
	}

	j, err := encodeToJSON(s.data, s.deadline)
	if err != nil {
		return err
	}

	expiry := s.deadline
	if s.opts.idleTimeout > 0 {
		ie := time.Now().Add(s.opts.idleTimeout)
		if ie.Before(expiry) {
			expiry = ie
		}
	}

	err = s.engine.Save(s.token, j, expiry)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    s.token,
		Path:     s.opts.path,
		Domain:   s.opts.domain,
		Secure:   s.opts.secure,
		HttpOnly: s.opts.httpOnly,
	}
	if s.opts.persist == true {
		cookie.Expires = expiry
		// The addition of 0.5 means MaxAge is correctly rounded to the nearest
		// second instead of being floored.
		cookie.MaxAge = int(expiry.Sub(time.Now()).Seconds() + 0.5)
	}
	http.SetCookie(w, cookie)
	s.written = true

	return nil
}

// RegenerateToken creates a new session token while retaining the current session
// data. The session lifetime is also reset.
//
// The old session token (and accompanying data) is deleted from the storage engine.
//
// To mitigate the risk of session fixation attacks, it's important that you call
// RegenerateToken before making any changes to privilege levels (e.g. login and
// logout operations). See https://www.owasp.org/index.php/Session_fixation for
// additional information.
//
// Usage:
//
//	func login(w http.ResponseWriter, r *http.Request) {
//		…
//		userID := 123
//		err := session.RegenerateToken(r)
//		if err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//		if err := session.PutInt(r, "user.id", userID); err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//		…
//	}
func RegenerateToken(r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}

	err = s.engine.Delete(s.token)
	if err != nil {
		return err
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	s.token = token
	s.deadline = time.Now().Add(s.opts.lifetime)
	s.modified = true

	return nil
}

// Renew creates a new session token and removes all data for the session. The
// session lifetime is also reset.
//
// The old session token (and accompanying data) is deleted from the storage engine.
//
// The Renew function is essentially a concurrency-safe amalgamation of the
// RegenerateToken and Clear functions.
func Renew(r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}

	err = s.engine.Delete(s.token)
	if err != nil {
		return err
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	s.token = token
	for key := range s.data {
		delete(s.data, key)
	}
	s.deadline = time.Now().Add(s.opts.lifetime)
	s.modified = true

	return nil
}

// Destroy deletes the current session. The session token (and any accompanying
// data) is deleted from the storage engine and the client is instructed to
// delete the session cookie.
//
// Destroy operations are effective immediately, and any future operations on
// the session within the same request cycle will return a ErrAlreadyWritten error
// (if you see this error, you probably want to use the Renew function instead).
//
// A new empty session will be created for any client that subsequently tries
// to use the destroyed session token.
func Destroy(w http.ResponseWriter, r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}

	err = s.engine.Delete(s.token)
	if err != nil {
		return err
	}

	s.token = ""
	for key := range s.data {
		delete(s.data, key)
	}
	s.modified = true

	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     s.opts.path,
		Domain:   s.opts.domain,
		Secure:   s.opts.secure,
		HttpOnly: s.opts.httpOnly,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	s.written = true

	return nil
}

func sessionFromContext(r *http.Request) (*session, error) {
	s, ok := r.Context().Value(ContextName).(*session)
	if ok == false {
		return nil, errors.New("request.Context does not contain a *session value")
	}
	return s, nil
}

func requestWithSession(r *http.Request, s *session) *http.Request {
	ctx := context.WithValue(r.Context(), ContextName, s)
	return r.WithContext(ctx)
}

func encodeToJSON(data map[string]interface{}, deadline time.Time) ([]byte, error) {
	return json.Marshal(&struct {
		Data     map[string]interface{} `json:"data"`
		Deadline int64                  `json:"deadline"`
	}{
		Data:     data,
		Deadline: deadline.UnixNano(),
	})
}

func decodeFromJSON(j []byte) (map[string]interface{}, time.Time, error) {
	aux := struct {
		Data     map[string]interface{} `json:"data"`
		Deadline int64                  `json:"deadline"`
	}{}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()
	err := dec.Decode(&aux)
	if err != nil {
		return nil, time.Time{}, err
	}
	return aux.Data, time.Unix(0, aux.Deadline), nil
}
