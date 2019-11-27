package scs

import (
	"bytes"
	"encoding/gob"
	"time"
)

// Codec is the interface for encoding/decoding session data to and from a byte
// slice for use by the session store.
type Codec interface {
	Encode(deadline time.Time, persit bool, values map[string]interface{}) ([]byte, error)
	Decode([]byte) (deadline time.Time, persit bool, values map[string]interface{}, err error)
}

type gobCodec struct{}

func (gobCodec) Encode(deadline time.Time, persit bool, values map[string]interface{}) ([]byte, error) {
	aux := &struct {
		Deadline time.Time
		Persist  bool
		Values   map[string]interface{}
	}{
		Deadline: deadline,
		Persist:  persit,
		Values:   values,
	}

	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(&aux)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (gobCodec) Decode(b []byte) (time.Time, bool, map[string]interface{}, error) {
	aux := &struct {
		Deadline time.Time
		Persist  bool
		Values   map[string]interface{}
	}{}

	r := bytes.NewReader(b)
	err := gob.NewDecoder(r).Decode(&aux)
	if err != nil {
		return time.Time{}, false, nil, err
	}

	return aux.Deadline, aux.Persist, aux.Values, nil
}
