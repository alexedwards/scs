package session

import (
	"net/http"
	"time"
)

func NewMockRequest(r *http.Request) *http.Request {
	do := *defaultOptions
	s := &session{
		token:    "",
		data:     make(map[string]interface{}),
		deadline: time.Now().Add(do.lifetime),
		engine:   &mockEngine{},
		opts:     &do,
	}
	return requestWithSession(r, s)
}

type mockEngine struct{}

func (me *mockEngine) Find(token string) (b []byte, exists bool, err error) {
	return nil, false, nil
}

func (me *mockEngine) Save(token string, b []byte, expiry time.Time) error {
	return nil
}

func (me *mockEngine) Delete(token string) error {
	return nil
}
