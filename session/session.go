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
		token:  token,
		values: make(map[string]interface{}),
		engine: engine,
		opts:   opts,
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

	j, found, err := engine.FindValues(token)
	if err != nil {
		return nil, err
	}
	if found == false {
		return newSession(r, engine, opts)
	}

	values, err := decodeValuesFromJSON(j)
	if err != nil {
		return nil, err
	}

	s := &session{
		token:  token,
		values: values,
		engine: engine,
		opts:   opts,
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

	if s.modified == false && s.opts.alwaysSave == false {
		return nil
	}

	j, err := encodeValuesToJSON(s.values)
	if err != nil {
		return err
	}

	err = s.engine.Save(s.token, j, s.opts.maxAge)
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
		cookie.Expires = time.Now().Add(s.opts.maxAge)
		cookie.MaxAge = int(s.opts.maxAge.Seconds())
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

func encodeValuesToJSON(values map[string]interface{}) ([]byte, error) {
	return json.Marshal(values)
}

func decodeValuesFromJSON(j []byte) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()
	err := dec.Decode(&values)
	if err != nil {
		return nil, err
	}
	return values, nil
}
