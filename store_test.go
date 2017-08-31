package scs

import (
	"errors"
	"time"
)

type mockStore struct {
	m map[string]*mockEntry
}

type mockEntry struct {
	b      []byte
	expiry time.Time
}

func newMockStore() *mockStore {
	m := make(map[string]*mockEntry)
	return &mockStore{m}
}

func (s *mockStore) Delete(token string) error {
	delete(s.m, token)
	return nil
}

func (s *mockStore) Find(token string) (b []byte, found bool, err error) {
	if token == "force-error" {
		return nil, false, errors.New("forced-error")
	}
	entry, exists := s.m[token]
	if !exists || entry.expiry.UnixNano() < time.Now().UnixNano() {
		return nil, false, nil
	}
	return entry.b, true, nil
}

func (s *mockStore) Save(token string, b []byte, expiry time.Time) error {
	s.m[token] = &mockEntry{b, expiry}
	return nil
}
