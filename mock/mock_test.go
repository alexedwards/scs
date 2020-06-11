package mock_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	testifyMock "github.com/stretchr/testify/mock"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/mock"
)

var (
	expectedCtx         = context.Background()
	expectedErr         = errors.New("wrong")
	expectedString      = "foo"
	expectedInt         = 1234
	expectedFloat       = 1234.1234
	expectedSliceString = []string{"foo", "bar"}
	expectedTime        = time.Now()
	expectedBool        = true
	expectedStatus      = scs.Status(1234)
	expectedBytes       = []byte{1, 2, 3, 4}
	expectedHTTPHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
)

func TestMock_Load(t *testing.T) {
	m := &mock.Mock{}
	m.On("Load", expectedCtx, expectedString).Return(expectedCtx, expectedErr)
	ctx, err := m.Load(expectedCtx, expectedString)
	assertValue(t, expectedErr, err)
	assertValue(t, expectedCtx, ctx)
	assertValue(t, expectedCtx, ctx)
	assertMock(t, &m.Mock)
}

func TestMock_Commit(t *testing.T) {
	m := &mock.Mock{}
	m.On("Commit", expectedCtx).Return(expectedString, expectedTime, expectedErr)
	s, ti, err := m.Commit(expectedCtx)
	assertValue(t, expectedErr, err)
	assertValue(t, expectedString, s)
	assertValue(t, expectedTime, ti)
	assertMock(t, &m.Mock)
}

func TestMock_Destroy(t *testing.T) {
	m := &mock.Mock{}
	m.On("Destroy", expectedCtx).Return(expectedErr)
	err := m.Destroy(expectedCtx)
	assertValue(t, expectedErr, err)
	assertMock(t, &m.Mock)
}

func TestMock_Put(t *testing.T) {
	m := &mock.Mock{}
	m.On("Put", expectedCtx, expectedString, "random")
	m.Put(expectedCtx, expectedString, "random")
	assertMock(t, &m.Mock)
}

func TestMock_Get(t *testing.T) {
	m := &mock.Mock{}
	m.On("Get", expectedCtx, expectedString).Return(expectedString)
	result := m.Get(expectedCtx, expectedString)
	assertValue(t, expectedString, result)
	assertMock(t, &m.Mock)
}

func TestMock_Pop(t *testing.T) {
	m := &mock.Mock{}
	m.On("Pop", expectedCtx, expectedString).Return(expectedString)
	result := m.Pop(expectedCtx, expectedString)
	assertValue(t, expectedString, result)
	assertMock(t, &m.Mock)
}

func TestMock_Remove(t *testing.T) {
	m := &mock.Mock{}
	m.On("Remove", expectedCtx, expectedString)
	m.Remove(expectedCtx, expectedString)
	assertMock(t, &m.Mock)
}

func TestMock_Clear(t *testing.T) {
	m := &mock.Mock{}
	m.On("Clear", expectedCtx).Return(expectedErr)
	err := m.Clear(expectedCtx)
	assertValue(t, expectedErr, err)
	assertMock(t, &m.Mock)
}

func TestMock_Exists(t *testing.T) {
	m := &mock.Mock{}
	m.On("Exists", expectedCtx, expectedString).Return(expectedBool)
	result := m.Exists(expectedCtx, expectedString)
	assertValue(t, expectedBool, result)
	assertMock(t, &m.Mock)
}

func TestMock_Keys(t *testing.T) {
	m := &mock.Mock{}
	m.On("Keys", expectedCtx).Return(expectedSliceString)
	result := m.Keys(expectedCtx)
	assertValue(t, expectedSliceString, result)
	assertMock(t, &m.Mock)
}

func TestMock_RenewToken(t *testing.T) {
	m := &mock.Mock{}
	m.On("RenewToken", expectedCtx).Return(expectedErr)
	result := m.RenewToken(expectedCtx)
	assertValue(t, expectedErr, result)
	assertMock(t, &m.Mock)
}

func TestMock_Status(t *testing.T) {
	m := &mock.Mock{}
	m.On("Status", expectedCtx).Return(expectedStatus)
	result := m.Status(expectedCtx)
	assertValue(t, expectedStatus, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetString(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetString", expectedCtx, expectedString).Return(expectedString)
	result := m.GetString(expectedCtx, expectedString)
	assertValue(t, expectedString, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetBool(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetBool", expectedCtx, expectedString).Return(expectedBool)
	result := m.GetBool(expectedCtx, expectedString)
	assertValue(t, expectedBool, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetInt(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetInt", expectedCtx, expectedString).Return(expectedInt)
	result := m.GetInt(expectedCtx, expectedString)
	assertValue(t, expectedInt, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetFloat(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetFloat", expectedCtx, expectedString).Return(expectedFloat)
	result := m.GetFloat(expectedCtx, expectedString)
	assertValue(t, expectedFloat, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetBytes(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetBytes", expectedCtx, expectedString).Return(expectedBytes)
	result := m.GetBytes(expectedCtx, expectedString)
	assertValue(t, expectedBytes, result)
	assertMock(t, &m.Mock)
}

func TestMock_GetTime(t *testing.T) {
	m := &mock.Mock{}
	m.On("GetTime", expectedCtx, expectedString).Return(expectedTime)
	result := m.GetTime(expectedCtx, expectedString)
	assertValue(t, expectedTime, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopString(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopString", expectedCtx, expectedString).Return(expectedString)
	result := m.PopString(expectedCtx, expectedString)
	assertValue(t, expectedString, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopBool(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopBool", expectedCtx, expectedString).Return(expectedBool)
	result := m.PopBool(expectedCtx, expectedString)
	assertValue(t, expectedBool, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopInt(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopInt", expectedCtx, expectedString).Return(expectedInt)
	result := m.PopInt(expectedCtx, expectedString)
	assertValue(t, expectedInt, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopFloat(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopFloat", expectedCtx, expectedString).Return(expectedFloat)
	result := m.PopFloat(expectedCtx, expectedString)
	assertValue(t, expectedFloat, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopBytes(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopBytes", expectedCtx, expectedString).Return(expectedBytes)
	result := m.PopBytes(expectedCtx, expectedString)
	assertValue(t, expectedBytes, result)
	assertMock(t, &m.Mock)
}

func TestMock_PopTime(t *testing.T) {
	m := &mock.Mock{}
	m.On("PopTime", expectedCtx, expectedString).Return(expectedTime)
	result := m.PopTime(expectedCtx, expectedString)
	assertValue(t, expectedTime, result)
	assertMock(t, &m.Mock)
}

func TestMock_LoadAndSave(t *testing.T) {
	m := &mock.Mock{}
	m.On("LoadAndSave", testifyMock.AnythingOfType("HandlerFunc")).Return(expectedHTTPHandler)
	result := m.LoadAndSave(expectedHTTPHandler)
	f1 := reflect.ValueOf(expectedHTTPHandler)
	f2 := reflect.ValueOf(result)
	if f1.Pointer() != f2.Pointer() {
		t.Errorf("want %+v; got %+v", f1.Pointer(), f2.Pointer())
	}
	assertMock(t, &m.Mock)
}

func assertValue(t *testing.T, expected, current interface{}) {
	if !reflect.DeepEqual(expected, current) {
		t.Errorf("want %+v; got %+v", expected, current)
	}
}

func assertMock(t *testing.T, m *testifyMock.Mock) {
	if !m.AssertExpectations(t) {
		t.Error("AssertMockFullFilled failed")
	}
}
