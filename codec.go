package scs

import (
	"bytes"
	"encoding/gob"
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

// Encode converts a session dealine and values into a byte slice.
func (GobCodec) Encode(deadline time.Time, values map[string]interface{}) ([]byte, error) {
	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{
		Deadline: deadline,
		Values:   values,
	}

	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(&aux)
	if err != nil {
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
	err := gob.NewDecoder(r).Decode(&aux)
	if err != nil {
		return time.Time{}, nil, err
	}

	return aux.Deadline, aux.Values, nil
}
