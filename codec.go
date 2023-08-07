package scs

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"time"
)

// Codec is the interface for encoding/decoding session data to and from a byte
// slice for use by the session store.
type Codec interface {
	Encode(deadline time.Time, values map[string]interface{}) ([]byte, error)
	Decode([]byte) (deadline time.Time, values map[string]interface{}, err error)
}

// GobCodec is used for encoding/decoding session data to and from a byte
// slice using the encoding/gob package.
type GobCodec struct{}

// Encode converts a session deadline and values into a byte slice.
func (GobCodec) Encode(deadline time.Time, values map[string]interface{}) ([]byte, error) {
	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{
		Deadline: deadline,
		Values:   values,
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(&aux); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Decode converts a byte slice into a session deadline and values.
func (GobCodec) Decode(b []byte) (time.Time, map[string]interface{}, error) {
	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{}

	r := bytes.NewReader(b)
	if err := gob.NewDecoder(r).Decode(&aux); err != nil {
		return time.Time{}, nil, err
	}

	return aux.Deadline, aux.Values, nil
}

// RegisterType registers a type to be available for adding in session data and following encoding/decoding operations
func (s *SessionManager) registerType(value interface{}) {
	gob.Register(value)
	rt := reflect.TypeOf(value).String()
	s.gobTypes = append(s.gobTypes, rt)
}

// checkRegisteredType checks if type has been registered
func (s *SessionManager) checkRegisteredType(value interface{}) bool {
	rt := reflect.TypeOf(value).String()
	for _, t := range s.gobTypes {
		if rt == t {
			return true
		}
	}
	return false
}
