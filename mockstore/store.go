package mockstore

import (
	"bytes"
	"time"
)

type expectedDelete struct {
	inputToken string
	returnErr  error
}

type expectedFind struct {
	inputToken  string
	returnB     []byte
	returnFound bool
	returnErr   error
}

type expectedCommit struct {
	inputToken  string
	inputB      []byte
	inputExpiry time.Time
	returnErr   error
}

type MockStore struct {
	deleteExpectations []expectedDelete
	findExpectations   []expectedFind
	commitExpectations []expectedCommit
}

// Delete implements the Store interface
func (m *MockStore) ExpectDelete(token string, returnErr error) {
	m.deleteExpectations = append(m.deleteExpectations, expectedDelete{
		inputToken: token,
		returnErr:  returnErr,
	})
}

// Delete implements the Store interface
func (m *MockStore) Delete(token string) (err error) {
	var (
		indexToRemove    int
		expectationFound bool
	)
	for i, expectation := range m.deleteExpectations {
		if expectation.inputToken == token {
			indexToRemove = i
			expectationFound = true
			break
		}
	}
	if !expectationFound {
		panic("store.Delete called unexpectedly")
	}

	errToReturn := m.deleteExpectations[indexToRemove].returnErr
	m.deleteExpectations = m.deleteExpectations[:indexToRemove+copy(m.deleteExpectations[indexToRemove:], m.deleteExpectations[indexToRemove+1:])]

	return errToReturn
}

func (m *MockStore) ExpectFind(token string, b []byte, found bool, err error) {
	m.findExpectations = append(m.findExpectations, expectedFind{
		inputToken:  token,
		returnB:     b,
		returnFound: found,
		returnErr:   err,
	})
}

// Find implements the Store interface
func (m *MockStore) Find(token string) (b []byte, found bool, err error) {
	var (
		indexToRemove    int
		expectationFound bool
	)
	for i, expectation := range m.findExpectations {
		if expectation.inputToken == token {
			indexToRemove = i
			expectationFound = true
			break
		}
	}
	if !expectationFound {
		panic("store.Find called unexpectedly")
	}

	valueToReturn := m.findExpectations[indexToRemove]
	m.findExpectations = m.findExpectations[:indexToRemove+copy(m.findExpectations[indexToRemove:], m.findExpectations[indexToRemove+1:])]

	return valueToReturn.returnB, valueToReturn.returnFound, valueToReturn.returnErr
}

func (m *MockStore) ExpectCommit(token string, b []byte, expiry time.Time, err error) {
	m.commitExpectations = append(m.commitExpectations, expectedCommit{
		inputToken:  token,
		inputB:      b,
		inputExpiry: expiry,
		returnErr:   err,
	})
}

// Commit implements the Store interface
func (m *MockStore) Commit(token string, b []byte, expiry time.Time) (err error) {
	var (
		indexToRemove    int
		expectationFound bool
	)
	for i, expectation := range m.commitExpectations {
		if expectation.inputToken == token && bytes.Compare(expectation.inputB, b) == 0 && expectation.inputExpiry == expiry {
			indexToRemove = i
			expectationFound = true
			break
		}
	}
	if !expectationFound {
		panic("store.Commit called unexpectedly")
	}

	errToReturn := m.commitExpectations[indexToRemove].returnErr
	m.commitExpectations = m.commitExpectations[:indexToRemove+copy(m.commitExpectations[indexToRemove:], m.commitExpectations[indexToRemove+1:])]

	return errToReturn
}
