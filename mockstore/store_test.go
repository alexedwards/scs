package mockstore

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestMockStore_Delete(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"
		expectedErr := errors.New("arbitrary")

		s.ExpectDelete(exampleToken, expectedErr)

		if err := s.Delete(exampleToken); err != expectedErr {
			t.Error("expected error not returned")
		}
		if len(s.deleteExpectations) != 0 {
			t.Error("expectations left over after exhausting calls")
		}
	})

	T.Run("panics with unfound expectation", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic to occur")
			}
		}()

		if err := s.Delete(exampleToken); err != nil {
			t.Error("unexpected error returned")
		}
	})
}

func TestMockStore_Find(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"
		expectedBytes := []byte("hello, world!")
		expectedFound := true

		s.ExpectFind(exampleToken, expectedBytes, expectedFound, nil)

		actualBytes, actualFound, actualErr := s.Find(exampleToken)
		if !bytes.Equal(expectedBytes, actualBytes) {
			t.Error("returned bytes do not match expectation")
		}
		if expectedFound != actualFound {
			t.Error("returned found does not match expectation")
		}
		if actualErr != nil {
			t.Error("unexpected error returned")
		}
		if len(s.findExpectations) != 0 {
			t.Error("expectations left over after exhausting calls")
		}
	})

	T.Run("panics with unfound expectation", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic to occur")
			}
		}()

		_, _, actualErr := s.Find(exampleToken)
		if actualErr != nil {
			t.Error("unexpected error returned")
		}
	})
}

func TestMockStore_Commit(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"
		exampleBytes := []byte("hello, world!")
		exampleExpiry := time.Now().Add(time.Hour)
		expectedErr := errors.New("arbitrary")

		s.ExpectCommit(exampleToken, exampleBytes, exampleExpiry, expectedErr)

		if err := s.Commit(exampleToken, exampleBytes, exampleExpiry); err != expectedErr {
			t.Error("expected error not returned")
		}
		if len(s.commitExpectations) != 0 {
			t.Error("expectations left over after exhausting calls")
		}
	})

	T.Run("panics with unfound expectation", func(t *testing.T) {
		s := &MockStore{}

		exampleToken := "token"
		exampleBytes := []byte("hello, world!")
		exampleExpiry := time.Now().Add(time.Hour)

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic to occur")
			}
		}()

		if err := s.Commit(exampleToken, exampleBytes, exampleExpiry); err != nil {
			t.Error("unexpected error returned")
		}
	})
}
