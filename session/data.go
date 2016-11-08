package session

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// ErrTypeAssertionFailed is returned by operations on session data where the
// received value could not be type asserted or converted into the required type.
var ErrTypeAssertionFailed = errors.New("type assertion failed")

// GetString returns the string value for a given key from the session data. The
// zero value for a string ("") is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// string.
func GetString(r *http.Request, key string) (string, error) {
	v, exists, err := get(r, key)
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
func PutString(r *http.Request, key string, val string) error {
	return put(r, key, val)
}

// PopString removes the string value for a given key from the session data
// and returns it. The zero value for a string ("") is returned if the key does
// not exist. An ErrTypeAssertionFailed error is returned if the value could not
// be type asserted to a string.
func PopString(r *http.Request, key string) (string, error) {
	v, exists, err := pop(r, key)
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
func GetBool(r *http.Request, key string) (bool, error) {
	v, exists, err := get(r, key)
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
func PutBool(r *http.Request, key string, val bool) error {
	return put(r, key, val)
}

// PopBool removes the bool value for a given key from the session data and returns
// it. The zero value for a bool (false) is returned if the key does not exist.
// An ErrTypeAssertionFailed error is returned if the value could not be type
// asserted to a bool.
func PopBool(r *http.Request, key string) (bool, error) {
	v, exists, err := pop(r, key)
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
func GetInt(r *http.Request, key string) (int, error) {
	v, exists, err := get(r, key)
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
func PutInt(r *http.Request, key string, val int) error {
	return put(r, key, val)
}

// PopInt removes the int value for a given key from the session data and returns
// it. The zero value for an int (0) is returned if the key does not exist. An
// ErrTypeAssertionFailed error is returned if the value could not be type asserted
// or converted to a int.
func PopInt(r *http.Request, key string) (int, error) {
	v, exists, err := pop(r, key)
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

//

// GetInt64 returns the int64 value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist. An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a int64.
func GetInt64(r *http.Request, key string) (int64, error) {
	v, exists, err := get(r, key)
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
func PutInt64(r *http.Request, key string, val int64) error {
	return put(r, key, val)
}

// PopInt64 remvoes the int64 value for a given key from the session data
// and returns it. The zero value for an int (0) is returned if the key does not
// exist. An ErrTypeAssertionFailed error is returned if the value could not be
// type asserted or converted to a int64.
func PopInt64(r *http.Request, key string) (int64, error) {
	v, exists, err := pop(r, key)
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
func GetFloat(r *http.Request, key string) (float64, error) {
	v, exists, err := get(r, key)
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
func PutFloat(r *http.Request, key string, val float64) error {
	return put(r, key, val)
}

// PopFloat removes the float64 value for a given key from the session data
// and returns it. The zero value for an float (0) is returned if the key does
// not exist. An ErrTypeAssertionFailed error is returned if the value could not
// be type asserted or converted to a float64.
func PopFloat(r *http.Request, key string) (float64, error) {
	v, exists, err := pop(r, key)
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
func GetTime(r *http.Request, key string) (time.Time, error) {
	v, exists, err := get(r, key)
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
func PutTime(r *http.Request, key string, val time.Time) error {
	return put(r, key, val)
}

// PopTime removes the time.Time value for a given key from the session data
// and returns it. The zero value for a time.Time object is returned if the key
// does not exist (this can be checked for with the time.IsZero method). An ErrTypeAssertionFailed
// error is returned if the value could not be type asserted or converted to a
// time.Time.
func PopTime(r *http.Request, key string) (time.Time, error) {
	v, exists, err := pop(r, key)
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
func GetBytes(r *http.Request, key string) ([]byte, error) {
	v, exists, err := get(r, key)
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
func PutBytes(r *http.Request, key string, val []byte) error {
	if val == nil {
		return errors.New("value must not be nil")
	}

	return put(r, key, val)
}

// PopBytes removes the byte slice ([]byte) value for a given key from the session
// data and returns it. The zero value for a slice (nil) is returned if the key
// does not exist. An ErrTypeAssertionFailed error is returned if the value could
// not be type asserted or converted to a []byte.
func PopBytes(r *http.Request, key string) ([]byte, error) {
	v, exists, err := pop(r, key)
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

/*
GetObject reads the data for a given session key into an arbitrary object
(represented by the dst parameter). It should only be used to retrieve custom
data types that have been stored using PutObject. The object represented by dst
will remain unchanged if the key does not exist.

The dst parameter must be a pointer.

Usage:

	// Note that the fields on the custom type are all exported.
	type User struct {
	    Name  string
	    Email string
	}

	func getHandler(w http.ResponseWriter, r *http.Request) {
	    // Register the type with the encoding/gob package. Usually this would be
	    // done in an init() function.
	    gob.Register(User{})

	    // Initialise a pointer to a new, empty, custom object.
	    user := &User{}

	    // Read the custom object data from the session into the pointer.
	    err := session.GetObject(r, "user", user)
	    if err != nil {
	        http.Error(w, err.Error(), 500)
	        return
	    }
	    fmt.Fprintf(w, "Name: %s, Email: %s", user.Name, user.Email)
	}
*/
func GetObject(r *http.Request, key string, dst interface{}) error {
	b, err := GetBytes(r, key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	return gobDecode(b, dst)
}

/*
PutObject adds an arbitrary object and corresponding key to the the session data.
Any existing value for the key will be replaced.

The val parameter must be a pointer to your object.

PutObject is typically used to store custom data types. It encodes the object
into a gob and then into a base64-encoded string which is persisted by the
storage engine. This makes PutObject (and the accompanying GetObject and PopObject
functions) comparatively expensive operations.

Because gob encoding is used, the fields on custom types must be exported in
order to be persisted correctly. Custom data types must also be registered with
gob.Register before PutObject is called (see https://golang.org/pkg/encoding/gob/#Register).

Usage:

  type User struct {
      Name  string
      Email string
  }

  func putHandler(w http.ResponseWriter, r *http.Request) {
      // Register the type with the encoding/gob package. Usually this would be
      // done in an init() function.
      gob.Register(User{})

      // Initialise a pointer to a new custom object.
      user := &User{"Alice", "alice@example.com"}

      // Store the custom object in the session data. Important: you should pass in
      // a pointer to your object, not the value.
      err := session.PutObject(r, "user", user)
      if err != nil {
          http.Error(w, err.Error(), 500)
      }
  }
*/
func PutObject(r *http.Request, key string, val interface{}) error {
	if val == nil {
		return errors.New("value must not be nil")
	}

	b, err := gobEncode(val)
	if err != nil {
		return err
	}

	return PutBytes(r, key, b)
}

// PopObject removes the data for a given session key and reads it into a custom
// object (represented by the dst parameter). It should only be used to retrieve
// custom data types that have been stored using PutObject. The object represented
// by dst will remain unchanged if the key does not exist.
//
// The dst parameter must be a pointer.
func PopObject(r *http.Request, key string, dst interface{}) error {
	b, err := PopBytes(r, key)
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}

	return gobDecode(b, dst)
}

// Exists returns true if the given key is present in the session data.
func Exists(r *http.Request, key string) (bool, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return false, err
	}

	s.mu.Lock()
	_, exists := s.data[key]
	s.mu.Unlock()

	return exists, nil
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

	_, exists := s.data[key]
	if exists == false {
		return nil
	}

	delete(s.data, key)
	s.modified = true
	return nil
}

// Clear removes all data for the current session. The session token and lifetime
// are unaffected. If there is no data in the current session this operation is
// a no-op.
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

	if len(s.data) == 0 {
		return nil
	}

	for key := range s.data {
		delete(s.data, key)
	}
	s.modified = true
	return nil
}

func get(r *http.Request, key string) (interface{}, bool, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return nil, false, err
	}

	s.mu.Lock()
	v, exists := s.data[key]
	s.mu.Unlock()

	return v, exists, nil
}

func put(r *http.Request, key string, val interface{}) error {
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

func pop(r *http.Request, key string) (interface{}, bool, error) {
	s, err := sessionFromContext(r)
	if err != nil {
		return "", false, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.written == true {
		return "", false, ErrAlreadyWritten
	}
	v, exists := s.data[key]
	if exists == false {
		return nil, false, nil
	}

	delete(s.data, key)
	s.modified = true
	return v, true, nil
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
