/*
Package session provides session management middleware and helpers for
manipulating session data.

It should be installed alongside one of the storage engines from https://godoc.org/github.com/alexedwards/scs/engine.

For example:

    $ go get github.com/alexedwards/scs/session
    $ go get github.com/alexedwards/scs/engine/memstore

Basic use:

    package main

    import (
        "io"
        "net/http"

        "github.com/alexedwards/scs/engine/memstore"
        "github.com/alexedwards/scs/session"
    )

    func main() {
        // Initialise a new storage engine. Here we use the memstore package, but the principles
        // are the same no matter which back-end store you choose.
        engine := memstore.New(0)

        // Initialise the session manager middleware, passing in the storage engine as
        // the first parameter. This middleware will automatically handle loading and
        // saving of session data for you.
        sessionManager := session.Manage(engine)

        // Set up your HTTP handlers in the normal way.
        mux := http.NewServeMux()
        mux.HandleFunc("/put", putHandler)
        mux.HandleFunc("/get", getHandler)

        // Wrap your handlers with the session manager middleware.
        http.ListenAndServe(":4000", sessionManager(mux))
    }

    func putHandler(w http.ResponseWriter, r *http.Request) {
        // Use the PutString helper to store a new key and associated string value in
        // the session data. Helpers are also available for bool, int, int64, float,
        // time.Time and []byte data types.
        err := session.PutString(r, "message", "Hello from a session!")
        if err != nil {
            http.Error(w, err.Error(), 500)
        }
    }

    func getHandler(w http.ResponseWriter, r *http.Request) {
        // Use the GetString helper to retrieve the string value associated with a key.
        msg, err := session.GetString(r, "message")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        io.WriteString(w, msg)
    }
*/
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
)

// ErrAlreadyWritten is returned when an attempt is made to modify the session
// data after it has already been sent to the storage engine and client.
var ErrAlreadyWritten = errors.New("session already written to the engine and http.ResponseWriter")

type session struct {
	token    string
	data     map[string]interface{}
	deadline time.Time
	engine   Engine
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

func newSession(r *http.Request, engine Engine, opts *options) (*http.Request, error) {
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

func load(r *http.Request, engine Engine, opts *options) (*http.Request, error) {
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

	if ce, ok := s.engine.(cookieEngine); ok {
		s.token, err = ce.MakeToken(j, expiry)
		if err != nil {
			return err
		}
	} else {
		err = s.engine.Save(s.token, j, expiry)
		if err != nil {
			return err
		}
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

/*
RegenerateToken creates a new session token while retaining the current session
data. The session lifetime is also reset.

The old session token and accompanying data are deleted from the storage engine.

To mitigate the risk of session fixation attacks, it's important that you call
RegenerateToken before making any changes to privilege levels (e.g. login and
logout operations). See https://www.owasp.org/index.php/Session_fixation for
additional information.

Usage:

	func loginHandler(w http.ResponseWriter, r *http.Request) {
	    userID := 123

	    // First regenerate the session token…
	    err := session.RegenerateToken(r)
	    if err != nil {
	        http.Error(w, err.Error(), 500)
	        return
	    }

	    // Then make the privilege-level change.
	    err = session.PutInt(r, "userID", userID)
	    if err != nil {
	        http.Error(w, err.Error(), 500)
	        return
	    }
	}

*/
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
// The old session token and accompanying data are deleted from the storage engine.
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

// Destroy deletes the current session. The session token and accompanying
// data are deleted from the storage engine, and the client is instructed to
// delete the session cookie.
//
// Destroy operations are effective immediately, and any further operations on
// the session in the same request cycle will return an ErrAlreadyWritten error.
// If you see this error you probably want to use the Renew function instead.
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

// Save immediately writes the session cookie header to the ResponseWriter and
// saves the session data to the storage engine, if needed.
//
// Using Save is not normally necessary. The session middleware (which buffers
// all writes to the underlying connection) will automatically handle setting the
// cookie header and storing the data for you.
//
// However there may be instances where you wish to break out of this normal
// operation and (one way or another) write to the underlying connection before
// control is passed back to the session middleware. In these instances, where
// response headers have already been written, the middleware will be too late
// to set the cookie header. The solution is to manually call Save before performing
// any writes.
//
// An example is flushing data using the http.Flusher interface:
//
//	func flushingHandler(w http.ResponseWriter, r *http.Request) {
//	 	err := session.PutString(r, "foo", "bar")
//		if err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//		err = session.Save(w, r)
//		if err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//
//		fw, ok := w.(http.Flusher)
//		if !ok {
//			http.Error(w, "could not assert to http.Flusher", 500)
//			return
//		}
//		w.Write([]byte("This is some…"))
//		fw.Flush()
//		w.Write([]byte("flushed data"))
//	}
func Save(w http.ResponseWriter, r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	wr := s.written
	s.mu.Unlock()
	if wr == true {
		return ErrAlreadyWritten
	}

	return write(w, r)
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
