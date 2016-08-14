package session

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// ErrKeyNotFound is returned by operations on session data when the given
// key does not exist.
var ErrKeyNotFound = errors.New("key not found in session data")

// ErrTypeAssertionFailed is returned by operations on session data where the
// received value could not be type asserted or converted into the required type.
var ErrTypeAssertionFailed = errors.New("type assertion failed")

// GetString returns the string value for a given key from the session data. An
// ErrKeyNotFound error is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// string.
func GetString(r *http.Request, key string) (string, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return "", ErrKeyNotFound
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}

	return str, nil
}

// PutString adds a string value and corresponding key to the the session data.
// Any existing value for the key will be replaced.
func PutString(r *http.Request, key string, val string) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopString returns the string value for a given key from the session data
// and then removes it (both the key and value). An ErrKeyNotFound error is returned
// if the key does not exist. An ErrTypeAssertionFailed error is returned if the
// value could not be type asserted to a string.
func PopString(r *http.Request, key string) (string, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return "", ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return "", ErrKeyNotFound
	}

	str, ok := v.(string)
	if ok == false {
		return "", ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return str, nil
}

// GetBool returns the bool value for a given key from the session data. An ErrKeyNotFound
// error is returned if the key does not exist. An ErrTypeAssertionFailed error
// is returned if the value could not be type asserted to a bool.
func GetBool(r *http.Request, key string) (bool, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return false, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return false, ErrKeyNotFound
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}

	return b, nil
}

// PutBool adds a bool value and corresponding key to the session data. Any existing
// value for the key will be replaced.
func PutBool(r *http.Request, key string, val bool) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopBool returns the bool value for a given key from the session data
// and then removes it (both the key and value). An ErrKeyNotFound error is returned
// if the key does not exist. An ErrTypeAssertionFailed error is returned if the
// value could not be type asserted to a bool.
func PopBool(r *http.Request, key string) (bool, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return false, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return false, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return false, ErrKeyNotFound
	}

	b, ok := v.(bool)
	if ok == false {
		return false, ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return b, nil
}

// GetInt returns the int value for a given key from the session data. An ErrKeyNotFound
// error is returned if the key does not exist. An ErrTypeAssertionFailed error
// is returned if the value could not be type asserted or converted to a int.
func GetInt(r *http.Request, key string) (int, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return 0, ErrKeyNotFound
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
func PutInt(r *http.Request, key string, val int) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopInt returns the int value for a given key from the session data
// and then removes it (both the key and value). An ErrKeyNotFound error is returned
// if the key does not exist. An ErrTypeAssertionFailed error is returned if the
// value could not be type asserted or converted to a int.
func PopInt(r *http.Request, key string) (int, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.written == true {
		return 0, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return 0, ErrKeyNotFound
	}

	var i int
	switch v := v.(type) {
	case int:
		i = v
	case json.Number:
		i, err = strconv.Atoi(v.String())
		if err != nil {
			return 0, err
		}
	default:
		return 0, ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return i, nil
}

// GetFloat returns the float64 value for a given key from the session data. An
// ErrKeyNotFound error is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// float64.
func GetFloat(r *http.Request, key string) (float64, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return 0, ErrKeyNotFound
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
func PutFloat(r *http.Request, key string, val float64) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopFloat returns the float64 value for a given key from the session data
// and then removes it (both the key and value). An ErrKeyNotFound error is returned
// if the key does not exist. An ErrTypeAssertionFailed error is returned if the
// value could not be type asserted or converted to a float64.
func PopFloat(r *http.Request, key string) (float64, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.written == true {
		return 0, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return 0, ErrKeyNotFound
	}

	var f float64
	switch v := v.(type) {
	case float64:
		f = v
	case json.Number:
		f, err = v.Float64()
		if err != nil {
			return 0, err
		}
	default:
		return 0, ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return f, nil
}

// GetTime returns the time.Time value for a given key from the session data. An
// ErrKeyNotFound error is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// time.Time.
func GetTime(r *http.Request, key string) (time.Time, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return time.Time{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return time.Time{}, ErrKeyNotFound
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
func PutTime(r *http.Request, key string, val time.Time) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopTime returns the time.Time value for a given key from the session data
// and then removes it (both the key and value). An ErrKeyNotFound error is returned
// if the key does not exist. An ErrTypeAssertionFailed error is returned if the
// value could not be type asserted or converted to a time.Time.
func PopTime(r *http.Request, key string) (time.Time, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return time.Time{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.written == true {
		return time.Time{}, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return time.Time{}, ErrKeyNotFound
	}

	var t time.Time
	switch v := v.(type) {
	case time.Time:
		t = v
	case string:
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
	default:
		return time.Time{}, ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return t, nil
}

// GetBytes returns the byte slice ([]byte) value for a given key from the session
// data. An ErrKeyNotFound error is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to
// []byte.
func GetBytes(r *http.Request, key string) ([]byte, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	v, exists := s.data[key]
	if exists == false {
		return nil, ErrKeyNotFound
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
func PutBytes(r *http.Request, key string, val []byte) error {
	if val == nil {
		return errors.New("value must not be nil")
	}

	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	s.data[key] = val
	s.modified = true
	return nil
}

// PopBytes returns the byte slice ([]byte) value for a given key from the session
// data and then removes it (both the key and value). An ErrKeyNotFound error is
// returned if the key does not exist. An ErrTypeAssertionFailed error is returned
// if the value could not be type asserted or converted to a []byte.
func PopBytes(r *http.Request, key string) ([]byte, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.written == true {
		return nil, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return nil, ErrKeyNotFound
	}

	var b []byte
	switch v := v.(type) {
	case []byte:
		b = v
	case string:
		b, err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrTypeAssertionFailed
	}

	delete(s.data, key)
	s.modified = true
	return b, nil
}

// Remove deletes the given key and corresponding value from the session data.
// If the key is not present this operation is a no-op.
func Remove(r *http.Request, key string) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	delete(s.data, key)
	s.modified = true
	return nil
}

// Clear removes all data for the current session. The session token and lifetime
// are unaffected.
func Clear(r *http.Request) error {
	s, err := sessionFromContext(r)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return ErrAlreadyWritten
	}
	for key := range s.data {
		delete(s.data, key)
	}
	s.modified = true
	return nil
}
