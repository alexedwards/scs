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

var (
	ErrAlreadyWritten     = errors.New("session already written to the engine and http.ResponseWriter")
	ErrNoSessionInContext = errors.New("request.Context does not contain a *session value")
)

type session struct {
	token    string
	values   map[string]interface{}
	deadline time.Time
	engine   scs.Engine
	opts     *options
	modified bool
	written  bool
	mu       sync.RWMutex
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
		values:   make(map[string]interface{}),
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

	values, deadline, err := decodeDataFromJSON(j)
	if err != nil {
		return nil, err
	}

	s := &session{
		token:    token,
		values:   values,
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

	j, err := encodeDataToJSON(s.values, s.deadline)
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
	for key := range s.values {
		delete(s.values, key)
	}
	s.deadline = time.Now().Add(s.opts.lifetime)
	s.modified = true

	return nil
}

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
	for key := range s.values {
		delete(s.values, key)
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
	s, ok := r.Context().Value(ContextDataName).(*session)
	if ok == false {
		return nil, ErrNoSessionInContext
	}
	return s, nil
}

func requestWithSession(r *http.Request, s *session) *http.Request {
	ctx := context.WithValue(r.Context(), ContextDataName, s)
	return r.WithContext(ctx)
}

func encodeDataToJSON(values map[string]interface{}, deadline time.Time) ([]byte, error) {
	return json.Marshal(&struct {
		Values   map[string]interface{} `json:"values"`
		Deadline int64                  `json:"deadline"`
	}{
		Values:   values,
		Deadline: deadline.UnixNano(),
	})
}

func decodeDataFromJSON(j []byte) (map[string]interface{}, time.Time, error) {
	aux := struct {
		Values   map[string]interface{} `json:"values"`
		Deadline int64                  `json:"deadline"`
	}{}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()
	err := dec.Decode(&aux)
	if err != nil {
		return nil, time.Time{}, err
	}
	return aux.Values, time.Unix(0, aux.Deadline), nil
}
