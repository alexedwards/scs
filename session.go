package scs

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// sessionName is a custom type for the request context key.
type sessionName string

// ErrTypeAssertionFailed is returned by operations on session data where the
// received value could not be type asserted or converted into the required type.
var ErrTypeAssertionFailed = errors.New("type assertion failed")

// Session contains data for the current session.
type Session struct {
	token    string
	data     map[string]interface{}
	deadline time.Time
	store    Store
	opts     *options
	loadErr  error
	mu       sync.Mutex
}

// cookie wraps http.Cookie, adding SameSite support
type cookie struct {
	std      *http.Cookie // "stdlib cookie"
	sameSite string
}

func (c *cookie) String() string {
	v := c.std.String()
	if c.sameSite != "" {
		v = v + "; SameSite=" + c.sameSite
	}
	return v
}

func newSession(store Store, opts *options) *Session {
	return &Session{
		data:     make(map[string]interface{}),
		deadline: time.Now().Add(opts.lifetime),
		store:    store,
		opts:     opts,
	}
}

func load(r *http.Request, store Store, opts *options) *Session {
	// Check to see if there is an already loaded session in the request context.
	val := r.Context().Value(sessionName(opts.name))
	if val != nil {
		s, ok := val.(*Session)
		if !ok {
			return &Session{loadErr: fmt.Errorf("scs: can not assert %T to *Session", val)}
		}
		return s
	}

	cookie, err := r.Cookie(opts.name)
	if err == http.ErrNoCookie {
		return newSession(store, opts)
	} else if err != nil {
		return &Session{loadErr: err}
	}

	if cookie.Value == "" {
		return newSession(store, opts)
	}
	token := cookie.Value

	j, found, err := store.Find(token)
	if err != nil {
		return &Session{loadErr: err}
	}
	if found == false {
		return newSession(store, opts)
	}

	data, deadline, err := decodeFromJSON(j)
	if err != nil {
		return &Session{loadErr: err}
	}

	s := &Session{
		token:    token,
		data:     data,
		deadline: deadline,
		store:    store,
		opts:     opts,
	}

	return s
}

func (s *Session) write(w http.ResponseWriter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	if ce, ok := s.store.(cookieStore); ok {
		s.token, err = ce.MakeToken(j, expiry)
		if err != nil {
			return err
		}
	} else {
		if s.token == "" {
			s.token, err = generateToken()
			if err != nil {
				return err
			}
		}
		err = s.store.Save(s.token, j, expiry)
		if err != nil {
			return err
		}
	}

	cookie := &cookie{
		std: &http.Cookie{
			Name:     s.opts.name,
			Value:    s.token,
			Path:     s.opts.path,
			Domain:   s.opts.domain,
			Secure:   s.opts.secure,
			HttpOnly: s.opts.httpOnly,
		},
		sameSite: s.opts.sameSite,
	}
	if s.opts.persist == true {
		// Round up expiry time to the nearest second.
		cookie.std.Expires = time.Unix(expiry.Unix()+1, 0)
		cookie.std.MaxAge = int(expiry.Sub(time.Now()).Seconds() + 1)
	}

	// Overwrite any existing cookie header for the session...
	var set bool
	for i, h := range w.Header()["Set-Cookie"] {
		if strings.HasPrefix(h, fmt.Sprintf("%s=", s.opts.name)) {
			w.Header()["Set-Cookie"][i] = cookie.String()
			set = true
			break
		}
	}
	// Or set a new one if necessary.
	if !set {
		w.Header().Add("Set-Cookie", cookie.String())
	}

	return nil
}

// Token returns the token value that represents given session data.
// NOTE: The method returns the empty string if session hasn't yet been written to the store.
// If you're using the CookieStore this token will change each time the session is modified.
func (s *Session) Token() string {
	return s.token
}

// GetString returns the string value for a given key from the session data. The
// zero value for a string ("") is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// string.
func (s *Session) GetString(key string) (string, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return "", err
	}
	if exists == false {
		return "", nil
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}
	return str, nil
}

// PutString adds a string value and corresponding key to the the session data.
// Any existing value for the key will be replaced.
func (s *Session) PutString(w http.ResponseWriter, key string, val string) error {
	return s.put(w, key, val)
}

// PopString removes the string value for a given key from the session data
// and returns it. The zero value for a string ("") is returned if the key does
// not exist. An ErrTypeAssertionFailed error is returned if the value could not
// be type asserted to a string.
func (s *Session) PopString(w http.ResponseWriter, key string) (string, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return "", err
	}
	if exists == false {
		return "", nil
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}
	return str, nil
}

// GetBool returns the bool value for a given key from the session data. The
// zero value for a bool (false) is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted to a bool.
func (s *Session) GetBool(key string) (bool, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return false, err
	}
	if exists == false {
		return false, nil
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}
	return b, nil
}

// PutBool adds a bool value and corresponding key to the session data. Any existing
// value for the key will be replaced.
func (s *Session) PutBool(w http.ResponseWriter, key string, val bool) error {
	return s.put(w, key, val)
}

// PopBool removes the bool value for a given key from the session data and returns
// it. The zero value for a bool (false) is returned if the key does not exist.
// An ErrTypeAssertionFailed error is returned if the value could not be type
// asserted to a bool.
func (s *Session) PopBool(w http.ResponseWriter, key string) (bool, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return false, err
	}
	if exists == false {
		return false, nil
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}
	return b, nil
}

// GetInt returns the int value for a given key from the session data. The zero
// value for an int (0) is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a int.
func (s *Session) GetInt(key string) (int, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int:
		return v, nil
	case json.Number:
		return strconv.Atoi(v.String())
	}
	return 0, ErrTypeAssertionFailed
}

// PutInt adds an int value and corresponding key to the session data. Any existing
// value for the key will be replaced.
func (s *Session) PutInt(w http.ResponseWriter, key string, val int) error {
	return s.put(w, key, val)
}

// PopInt removes the int value for a given key from the session data and returns
// it. The zero value for an int (0) is returned if the key does not exist. An
// ErrTypeAssertionFailed error is returned if the value could not be type asserted
// or converted to a int.
func (s *Session) PopInt(w http.ResponseWriter, key string) (int, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int:
		return v, nil
	case json.Number:
		return strconv.Atoi(v.String())
	}
	return 0, ErrTypeAssertionFailed
}

// GetInt64 returns the int64 value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a int64.
func (s *Session) GetInt64(key string) (int64, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	}
	return 0, ErrTypeAssertionFailed
}

// PutInt64 adds an int64 value and corresponding key to the session data. Any existing
// value for the key will be replaced.
func (s *Session) PutInt64(w http.ResponseWriter, key string, val int64) error {
	return s.put(w, key, val)
}

// PopInt64 remvoes the int64 value for a given key from the session data
// and returns it. The zero value for an int (0) is returned if the key does not
// exist. An ErrTypeAssertionFailed error is returned if the value could not be
// type asserted or converted to a int64.
func (s *Session) PopInt64(w http.ResponseWriter, key string) (int64, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	}
	return 0, ErrTypeAssertionFailed
}

// GetFloat returns the float64 value for a given key from the session data. The
// zero value for an float (0) is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// float64.
func (s *Session) GetFloat(key string) (float64, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	}
	return 0, ErrTypeAssertionFailed
}

// PutFloat adds an float64 value and corresponding key to the session data. Any
// existing value for the key will be replaced.
func (s *Session) PutFloat(w http.ResponseWriter, key string, val float64) error {
	return s.put(w, key, val)
}

// PopFloat removes the float64 value for a given key from the session data
// and returns it. The zero value for an float (0) is returned if the key does
// not exist. An ErrTypeAssertionFailed error is returned if the value could not
// be type asserted or converted to a float64.
func (s *Session) PopFloat(w http.ResponseWriter, key string) (float64, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return 0, err
	}
	if exists == false {
		return 0, nil
	}

	switch v := v.(type) {
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	}
	return 0, ErrTypeAssertionFailed
}

// GetTime returns the time.Time value for a given key from the session data. The
// zero value for a time.Time object is returned if the key does not exist (this
// can be checked for with the time.IsZero method). An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// time.Time.
func (s *Session) GetTime(key string) (time.Time, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return time.Time{}, err
	}
	if exists == false {
		return time.Time{}, nil
	}

	switch v := v.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	}
	return time.Time{}, ErrTypeAssertionFailed
}

// PutTime adds an time.Time value and corresponding key to the session data. Any
// existing value for the key will be replaced.
func (s *Session) PutTime(w http.ResponseWriter, key string, val time.Time) error {
	return s.put(w, key, val)
}

// PopTime removes the time.Time value for a given key from the session data
// and returns it. The zero value for a time.Time object is returned if the key
// does not exist (this can be checked for with the time.IsZero method). An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// time.Time.
func (s *Session) PopTime(w http.ResponseWriter, key string) (time.Time, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return time.Time{}, err
	}
	if exists == false {
		return time.Time{}, nil
	}

	switch v := v.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	}
	return time.Time{}, ErrTypeAssertionFailed
}

// GetBytes returns the byte slice ([]byte) value for a given key from the session
// data. The zero value for a slice (nil) is returned if the key does not exist.
// An ErrTypeAssertionFailed error is returned if the value could not be type
// asserted or converted to []byte.
func (s *Session) GetBytes(key string) ([]byte, error) {
	v, exists, err := s.get(key)
	if err != nil {
		return nil, err
	}
	if exists == false {
		return nil, nil
	}

	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return base64.StdEncoding.DecodeString(v)
	}
	return nil, ErrTypeAssertionFailed
}

// PutBytes adds a byte slice ([]byte) value and corresponding key to the the
// session data. Any existing value for the key will be replaced.
func (s *Session) PutBytes(w http.ResponseWriter, key string, val []byte) error {
	if val == nil {
		return errors.New("value must not be nil")
	}

	return s.put(w, key, val)
}

// PopBytes removes the byte slice ([]byte) value for a given key from the session
// data and returns it. The zero value for a slice (nil) is returned if the key
// does not exist. An ErrTypeAssertionFailed error is returned if the value could
// not be type asserted or converted to a []byte.
func (s *Session) PopBytes(w http.ResponseWriter, key string) ([]byte, error) {
	v, exists, err := s.pop(w, key)
	if err != nil {
		return nil, err
	}
	if exists == false {
		return nil, nil
	}

	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return base64.StdEncoding.DecodeString(v)
	}
	return nil, ErrTypeAssertionFailed
}

// GetObject reads the data for a given session key into an arbitrary object
// (represented by the dst parameter). It should only be used to retrieve custom
// data types that have been stored using PutObject. The object represented by dst
// will remain unchanged if the key does not exist.
//
// The dst parameter must be a pointer.
func (s *Session) GetObject(key string, dst interface{}) error {
	b, err := s.GetBytes(key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	return gobDecode(b, dst)
}

// PutObject adds an arbitrary object and corresponding key to the the session data.
// Any existing value for the key will be replaced.
//
// The val parameter must be a pointer to your object.
//
// PutObject is typically used to store custom data types. It encodes the object
// into a gob and then into a base64-encoded string which is persisted by the
// session store. This makes PutObject (and the accompanying GetObject and PopObject
// functions) comparatively expensive operations.
//
// Because gob encoding is used, the fields on custom types must be exported in
// order to be persisted correctly. Custom data types must also be registered with
// gob.Register before PutObject is called (see https://golang.org/pkg/encoding/gob/#Register).
func (s *Session) PutObject(w http.ResponseWriter, key string, val interface{}) error {
	if val == nil {
		return errors.New("value must not be nil")
	}

	b, err := gobEncode(val)
	if err != nil {
		return err
	}

	return s.PutBytes(w, key, b)
}

// PopObject removes the data for a given session key and reads it into a custom
// object (represented by the dst parameter). It should only be used to retrieve
// custom data types that have been stored using PutObject. The object represented
// by dst will remain unchanged if the key does not exist.
//
// The dst parameter must be a pointer.
func (s *Session) PopObject(w http.ResponseWriter, key string, dst interface{}) error {
	b, err := s.PopBytes(w, key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	return gobDecode(b, dst)
}

// Keys returns a slice of all key names present in the session data, sorted
// alphabetically. If the session contains no data then an empty slice will be
// returned.
func (s *Session) Keys() ([]string, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, len(s.data))
	i := 0
	for k := range s.data {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	return keys, nil
}

// Exists returns true if the given key is present in the session data.
func (s *Session) Exists(key string) (bool, error) {
	if s.loadErr != nil {
		return false, s.loadErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.data[key]
	return exists, nil
}

// Remove deletes the given key and corresponding value from the session data.
// If the key is not present this operation is a no-op.
func (s *Session) Remove(w http.ResponseWriter, key string) error {
	if s.loadErr != nil {
		return s.loadErr
	}

	s.mu.Lock()

	_, exists := s.data[key]
	if exists == false {
		s.mu.Unlock()
		return nil
	}

	delete(s.data, key)
	s.mu.Unlock()

	return s.write(w)
}

// Clear removes all data for the current session. The session token and lifetime
// are unaffected. If there is no data in the current session this operation is
// a no-op.
func (s *Session) Clear(w http.ResponseWriter) error {
	if s.loadErr != nil {
		return s.loadErr
	}

	s.mu.Lock()

	if len(s.data) == 0 {
		s.mu.Unlock()
		return nil
	}

	for key := range s.data {
		delete(s.data, key)
	}
	s.mu.Unlock()

	return s.write(w)
}

// RenewToken creates a new session token while retaining the current session
// data. The session lifetime is also reset.
//
// The old session token and accompanying data are deleted from the session store.
//
// To mitigate the risk of session fixation attacks, it's important that you call
// RenewToken before making any changes to privilege levels (e.g. login and
// logout operations). See https://www.owasp.org/index.php/Session_fixation for
// additional information.
func (s *Session) RenewToken(w http.ResponseWriter) error {
	if s.loadErr != nil {
		return s.loadErr
	}

	s.mu.Lock()

	err := s.store.Delete(s.token)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	token, err := generateToken()
	if err != nil {
		s.mu.Unlock()
		return err
	}

	s.token = token
	s.deadline = time.Now().Add(s.opts.lifetime)
	s.mu.Unlock()

	return s.write(w)
}

// Destroy deletes the current session. The session token and accompanying
// data are deleted from the session store, and the client is instructed to
// delete the session cookie.
//
// Any further operations on the session in the same request cycle will result in a
// new session being created.
//
// A new empty session will be created for any client that subsequently tries
// to use the destroyed session token.
func (s *Session) Destroy(w http.ResponseWriter) error {
	if s.loadErr != nil {
		return s.loadErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.store.Delete(s.token)
	if err != nil {
		return err
	}

	s.token = ""
	for key := range s.data {
		delete(s.data, key)
	}

	cookie := &http.Cookie{
		Name:     s.opts.name,
		Value:    "",
		Path:     s.opts.path,
		Domain:   s.opts.domain,
		Secure:   s.opts.secure,
		HttpOnly: s.opts.httpOnly,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	return nil
}

// Touch writes the session data in order to update the expiry time when an
// Idle Timeout has been set. If IdleTimeout is not > 0, then Touch is a no-op.
func (s *Session) Touch(w http.ResponseWriter) error {
	if s.loadErr != nil {
		return s.loadErr
	}
	if s.opts.idleTimeout > 0 {
		return s.write(w)
	}
	return nil
}

func (s *Session) get(key string) (interface{}, bool, error) {
	if s.loadErr != nil {
		return nil, false, s.loadErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.data[key]
	return v, exists, nil
}

func (s *Session) put(w http.ResponseWriter, key string, val interface{}) error {
	if s.loadErr != nil {
		return s.loadErr
	}

	s.mu.Lock()
	s.data[key] = val
	s.mu.Unlock()

	return s.write(w)
}

func (s *Session) pop(w http.ResponseWriter, key string) (interface{}, bool, error) {
	if s.loadErr != nil {
		return nil, false, s.loadErr
	}

	s.mu.Lock()

	v, exists := s.data[key]
	if exists == false {
		s.mu.Unlock()
		return nil, false, nil
	}

	delete(s.data, key)
	s.mu.Unlock()

	err := s.write(w)
	if err != nil {
		return nil, false, err
	}

	return v, true, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func gobEncode(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(b []byte, dst interface{}) error {
	buf := bytes.NewBuffer(b)
	return gob.NewDecoder(buf).Decode(dst)
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
