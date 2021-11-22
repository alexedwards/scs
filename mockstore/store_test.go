package mockstore

import (
	"bytes"
	"errors"
	"reflect"
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
		if !bytes.Equal(actualBytes, expectedBytes) {
			t.Error("returned bytes do not match expectation")
		}
		if actualFound != expectedFound {
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

func TestMockStore_All(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		s := &MockStore{}

		expectedMapBytes := map[string][]byte{"token1": []byte("hello, world 1!"), "token2": []byte("hello, world 2!"), "token3": []byte("hello, world 3!")}

		s.ExpectAll(expectedMapBytes, nil)

		actualMapBytes, actualErr := s.All()
		if !reflect.DeepEqual(actualMapBytes, expectedMapBytes) {
			t.Error("returned map bytes do not match expectation")
		}
		if actualErr != nil {
			t.Error("unexpected error returned")
		}
		if len(s.allExpectations) != 0 {
			t.Error("expectations left over after exhausting calls")
		}
	})

	T.Run("panics with unfound expectation", func(t *testing.T) {
		s := &MockStore{}

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic to occur")
			}
		}()

		_, actualErr := s.All()
		if actualErr != nil {
			t.Error("unexpected error returned")
		}
	})
}
