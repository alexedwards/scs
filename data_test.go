package scs

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2/mockstore"
)

func TestSessionDataFromContext(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	s := New()
	s.getSessionDataFromContext(context.Background())
}

func TestSessionManager_Load(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := New()
		s.IdleTimeout = time.Hour * 24

		ctx := context.Background()
		expected := "example"
		exampleDeadline := time.Now().Add(time.Hour)

		encodedValue, err := s.Codec.Encode(exampleDeadline, map[string]interface{}{
			"things": "stuff",
		})
		if err != nil {
			t.Errorf("unexpected error encoding value: %v", err)
		}

		if err := s.Store.Commit(expected, encodedValue, exampleDeadline); err != nil {
			t.Errorf("error committing to session store: %v", err)
		}

		newCtx, err := s.Load(ctx, expected)
		if err != nil {
			t.Errorf("error loading from session manager: %v", err)
		}
		if newCtx == nil {
			t.Error("returned context is unexpectedly nil")
		}

		sd, ok := newCtx.Value(s.contextKey).(*sessionData)
		if !ok {
			t.Error("sessionData not present in returned context")
		}
		if sd == nil {
			t.Error("sessionData present in returned context unexpectedly nil")
			return
		}

		actual := sd.token

		if expected != actual {
			t.Errorf("expected %s to equal %s", expected, actual)
		}
	})

	T.Run("with preexisting session data", func(t *testing.T) {
		s := New()

		obligatorySessionData := &sessionData{}
		ctx := context.WithValue(context.Background(), s.contextKey, obligatorySessionData)
		expected := "example"

		newCtx, err := s.Load(ctx, expected)
		if err != nil {
			t.Errorf("error loading from session manager: %v", err)
		}
		if newCtx == nil {
			t.Error("returned context is unexpectedly nil")
		}
	})

	T.Run("with empty token", func(t *testing.T) {
		s := New()

		ctx := context.Background()
		expected := ""
		exampleDeadline := time.Now().Add(time.Hour)

		encodedValue, err := s.Codec.Encode(exampleDeadline, map[string]interface{}{
			"things": "stuff",
		})
		if err != nil {
			t.Errorf("unexpected error encoding value: %v", err)
		}

		if err := s.Store.Commit(expected, encodedValue, exampleDeadline); err != nil {
			t.Errorf("error committing to session store: %v", err)
		}

		newCtx, err := s.Load(ctx, "")
		if err != nil {
			t.Errorf("error loading from session manager: %v", err)
		}
		if newCtx == nil {
			t.Error("returned context is unexpectedly nil")
		}

		sd, ok := newCtx.Value(s.contextKey).(*sessionData)
		if !ok {
			t.Error("sessionData not present in returned context")
		}
		if sd == nil {
			t.Error("sessionData present in returned context unexpectedly nil")
			return
		}

		actual := sd.token

		if expected != actual {
			t.Errorf("expected %s to equal %s", expected, actual)
		}
	})

	T.Run("with error finding token in store", func(t *testing.T) {
		s := New()
		store := &mockstore.MockStore{}

		ctx := context.Background()
		expected := "example"

		store.ExpectFind(expected, []byte{}, true, errors.New("arbitrary"))
		s.Store = store

		newCtx, err := s.Load(ctx, expected)
		if err == nil {
			t.Errorf("no error loading from session manager: %v", err)
		}
		if newCtx != nil {
			t.Error("returned context is unexpectedly not nil")
		}
	})

	T.Run("with unfound token in store", func(t *testing.T) {
		s := New()

		ctx := context.Background()
		exampleToken := "example"
		expected := ""

		newCtx, err := s.Load(ctx, exampleToken)
		if err != nil {
			t.Errorf("error loading from session manager: %v", err)
		}
		if newCtx == nil {
			t.Error("returned context is unexpectedly nil")
		}

		sd, ok := newCtx.Value(s.contextKey).(*sessionData)
		if !ok {
			t.Error("sessionData not present in returned context")
		}
		if sd == nil {
			t.Error("sessionData present in returned context unexpectedly nil")
			return
		}

		actual := sd.token

		if expected != actual {
			t.Errorf("expected %s to equal %s", expected, actual)
		}
	})

	T.Run("with error decoding found token", func(t *testing.T) {
		s := New()

		ctx := context.Background()
		expected := "example"
		exampleDeadline := time.Now().Add(time.Hour)

		if err := s.Store.Commit(expected, []byte(""), exampleDeadline); err != nil {
			t.Errorf("error committing to session store: %v", err)
		}

		newCtx, err := s.Load(ctx, expected)
		if err == nil {
			t.Errorf("no error loading from session manager: %v", err)
		}
		if newCtx != nil {
			t.Error("returned context is unexpectedly nil")
		}
	})

	T.Run("with token hashing", func(t *testing.T) {
		s := New()
		s.HashTokenInStore = true
		s.IdleTimeout = time.Hour * 24

		expectedToken := "example"
		expectedExpiry := time.Now().Add(time.Hour)

		initialCtx := context.WithValue(context.Background(), s.contextKey, &sessionData{
			deadline: expectedExpiry,
			token:    expectedToken,
			values: map[string]interface{}{
				"blah": "blah",
			},
		})

		actualToken, actualExpiry, err := s.Commit(initialCtx)
		if expectedToken != actualToken {
			t.Errorf("expected token to equal %q, but received %q", expectedToken, actualToken)
		}
		if expectedExpiry != actualExpiry {
			t.Errorf("expected expiry to equal %v, but received %v", expectedExpiry, actualExpiry)
		}
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}

		retrievedCtx, err := s.Load(context.Background(), expectedToken)
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
		retrievedSessionData, ok := retrievedCtx.Value(s.contextKey).(*sessionData)
		if !ok {
			t.Errorf("unexpected data in retrieved context")
		} else if retrievedSessionData.token != expectedToken {
			t.Errorf("expected token in context's session data data to equal %v, but received %v", expectedToken, retrievedSessionData.token)
		}

		if err := s.Destroy(retrievedCtx); err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
	})
}

func TestSessionManager_Commit(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := New()
		s.IdleTimeout = time.Hour * 24

		expectedToken := "example"
		expectedExpiry := time.Now().Add(time.Hour)

		ctx := context.WithValue(context.Background(), s.contextKey, &sessionData{
			deadline: expectedExpiry,
			token:    expectedToken,
			values: map[string]interface{}{
				"blah": "blah",
			},
		})

		actualToken, actualExpiry, err := s.Commit(ctx)
		if expectedToken != actualToken {
			t.Errorf("expected token to equal %q, but received %q", expectedToken, actualToken)
		}
		if expectedExpiry != actualExpiry {
			t.Errorf("expected expiry to equal %v, but received %v", expectedExpiry, actualExpiry)
		}
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
	})

	T.Run("with empty token", func(t *testing.T) {
		s := New()
		s.IdleTimeout = time.Hour * 24

		expectedToken := "XO6_D4NBpGP3D_BtekxTEO6o2ZvOzYnArauSQbgg"
		expectedExpiry := time.Now().Add(time.Hour)

		ctx := context.WithValue(context.Background(), s.contextKey, &sessionData{
			deadline: expectedExpiry,
			token:    expectedToken,
			values: map[string]interface{}{
				"blah": "blah",
			},
		})

		actualToken, actualExpiry, err := s.Commit(ctx)
		if expectedToken != actualToken {
			t.Errorf("expected token to equal %q, but received %q", expectedToken, actualToken)
		}
		if expectedExpiry != actualExpiry {
			t.Errorf("expected expiry to equal %v, but received %v", expectedExpiry, actualExpiry)
		}
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
	})

	T.Run("with expired deadline", func(t *testing.T) {
		s := New()
		s.IdleTimeout = time.Millisecond

		expectedToken := "example"
		expectedExpiry := time.Now().Add(time.Hour * -100)

		ctx := context.WithValue(context.Background(), s.contextKey, &sessionData{
			deadline: time.Now().Add(time.Hour * 24),
			token:    expectedToken,
			values: map[string]interface{}{
				"blah": "blah",
			},
		})

		actualToken, actualExpiry, err := s.Commit(ctx)
		if expectedToken != actualToken {
			t.Errorf("expected token to equal %q, but received %q", expectedToken, actualToken)
		}
		if expectedExpiry == actualExpiry {
			t.Errorf("expected expiry not to equal %v", actualExpiry)
		}
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
	})

	T.Run("with error committing to store", func(t *testing.T) {
		s := New()
		s.IdleTimeout = time.Hour * 24

		store := &mockstore.MockStore{}
		expectedErr := errors.New("arbitrary")

		sd := &sessionData{
			deadline: time.Now().Add(time.Hour),
			token:    "example",
			values: map[string]interface{}{
				"blah": "blah",
			},
		}
		expectedBytes, err := s.Codec.Encode(sd.deadline, sd.values)
		if err != nil {
			t.Errorf("unexpected encode error: %v", err)
		}

		ctx := context.WithValue(context.Background(), s.contextKey, sd)

		store.ExpectCommit(sd.token, expectedBytes, sd.deadline, expectedErr)
		s.Store = store

		actualToken, _, err := s.Commit(ctx)
		if actualToken != "" {
			t.Error("expected empty token")
		}
		if err == nil {
			t.Error("expected error not returned")
		}
	})

	T.Run("with token hashing", func(t *testing.T) {
		s := New()
		s.HashTokenInStore = true
		s.IdleTimeout = time.Hour * 24

		expectedToken := "example"
		expectedExpiry := time.Now().Add(time.Hour)

		ctx := context.WithValue(context.Background(), s.contextKey, &sessionData{
			deadline: expectedExpiry,
			token:    expectedToken,
			values: map[string]interface{}{
				"blah": "blah",
			},
		})

		actualToken, actualExpiry, err := s.Commit(ctx)
		if expectedToken != actualToken {
			t.Errorf("expected token to equal %q, but received %q", expectedToken, actualToken)
		}
		if expectedExpiry != actualExpiry {
			t.Errorf("expected expiry to equal %v, but received %v", expectedExpiry, actualExpiry)
		}
		if err != nil {
			t.Errorf("unexpected error returned: %v", err)
		}
	})
}

func TestPut(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	ctx := s.addSessionDataToContext(context.Background(), sd)

	s.Put(ctx, "foo", "bar")

	if sd.values["foo"] != "bar" {
		t.Errorf("got %q: expected %q", sd.values["foo"], "bar")
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str, ok := s.Get(ctx, "foo").(string)
	if !ok {
		t.Errorf("could not convert %T to string", s.Get(ctx, "foo"))
	}

	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}
}

func TestPop(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str, ok := s.Pop(ctx, "foo").(string)
	if !ok {
		t.Errorf("could not convert %T to string", s.Get(ctx, "foo"))
	}

	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}

	_, ok = sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	s.Remove(ctx, "foo")

	if sd.values["foo"] != nil {
		t.Errorf("got %v: expected %v", sd.values["foo"], nil)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestClear(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	sd.values["baz"] = "boz"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	if err := s.Clear(ctx); err != nil {
		t.Errorf("unexpected error encountered clearing session: %v", err)
	}

	if sd.values["foo"] != nil {
		t.Errorf("got %v: expected %v", sd.values["foo"], nil)
	}

	if sd.values["baz"] != nil {
		t.Errorf("got %v: expected %v", sd.values["baz"], nil)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	if !s.Exists(ctx, "foo") {
		t.Errorf("got %v: expected %v", s.Exists(ctx, "foo"), true)
	}

	if s.Exists(ctx, "baz") {
		t.Errorf("got %v: expected %v", s.Exists(ctx, "baz"), false)
	}
}

func TestKeys(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	sd.values["woo"] = "waa"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	keys := s.Keys(ctx)
	if !reflect.DeepEqual(keys, []string{"foo", "woo"}) {
		t.Errorf("got %v: expected %v", keys, []string{"foo", "woo"})
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str := s.GetString(ctx, "foo")
	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}

	str = s.GetString(ctx, "baz")
	if str != "" {
		t.Errorf("got %q: expected %q", str, "")
	}
}

func TestGetBool(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = true
	ctx := s.addSessionDataToContext(context.Background(), sd)

	b := s.GetBool(ctx, "foo")
	if b != true {
		t.Errorf("got %v: expected %v", b, true)
	}

	b = s.GetBool(ctx, "baz")
	if b != false {
		t.Errorf("got %v: expected %v", b, false)
	}
}

func TestGetInt(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = 123
	ctx := s.addSessionDataToContext(context.Background(), sd)

	i := s.GetInt(ctx, "foo")
	if i != 123 {
		t.Errorf("got %v: expected %d", i, 123)
	}

	i = s.GetInt(ctx, "baz")
	if i != 0 {
		t.Errorf("got %v: expected %d", i, 0)
	}
}

func TestGetFloat(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = 123.456
	ctx := s.addSessionDataToContext(context.Background(), sd)

	f := s.GetFloat(ctx, "foo")
	if f != 123.456 {
		t.Errorf("got %v: expected %f", f, 123.456)
	}

	f = s.GetFloat(ctx, "baz")
	if f != 0 {
		t.Errorf("got %v: expected %f", f, 0.00)
	}
}

func TestGetBytes(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = []byte("bar")
	ctx := s.addSessionDataToContext(context.Background(), sd)

	b := s.GetBytes(ctx, "foo")
	if !bytes.Equal(b, []byte("bar")) {
		t.Errorf("got %v: expected %v", b, []byte("bar"))
	}

	b = s.GetBytes(ctx, "baz")
	if b != nil {
		t.Errorf("got %v: expected %v", b, nil)
	}
}

func TestGetTime(t *testing.T) {
	t.Parallel()

	now := time.Now()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = now
	ctx := s.addSessionDataToContext(context.Background(), sd)

	tm := s.GetTime(ctx, "foo")
	if tm != now {
		t.Errorf("got %v: expected %v", tm, now)
	}

	tm = s.GetTime(ctx, "baz")
	if !tm.IsZero() {
		t.Errorf("got %v: expected %v", tm, time.Time{})
	}
}

func TestPopString(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str := s.PopString(ctx, "foo")
	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}

	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	str = s.PopString(ctx, "bar")
	if str != "" {
		t.Errorf("got %q: expected %q", str, "")
	}
}

func TestPopBool(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = true
	ctx := s.addSessionDataToContext(context.Background(), sd)

	b := s.PopBool(ctx, "foo")
	if b != true {
		t.Errorf("got %v: expected %v", b, true)
	}

	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	b = s.PopBool(ctx, "bar")
	if b != false {
		t.Errorf("got %v: expected %v", b, false)
	}
}

func TestPopInt(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = 123
	ctx := s.addSessionDataToContext(context.Background(), sd)

	i := s.PopInt(ctx, "foo")
	if i != 123 {
		t.Errorf("got %d: expected %d", i, 123)
	}

	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	i = s.PopInt(ctx, "bar")
	if i != 0 {
		t.Errorf("got %d: expected %d", i, 0)
	}
}

func TestPopFloat(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = 123.456
	ctx := s.addSessionDataToContext(context.Background(), sd)

	f := s.PopFloat(ctx, "foo")
	if f != 123.456 {
		t.Errorf("got %f: expected %f", f, 123.456)
	}

	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	f = s.PopFloat(ctx, "bar")
	if f != 0.0 {
		t.Errorf("got %f: expected %f", f, 0.0)
	}
}

func TestPopBytes(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = []byte("bar")
	ctx := s.addSessionDataToContext(context.Background(), sd)

	b := s.PopBytes(ctx, "foo")
	if !bytes.Equal(b, []byte("bar")) {
		t.Errorf("got %v: expected %v", b, []byte("bar"))
	}
	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	b = s.PopBytes(ctx, "bar")
	if b != nil {
		t.Errorf("got %v: expected %v", b, nil)
	}
}

func TestPopTime(t *testing.T) {
	t.Parallel()

	now := time.Now()
	s := New()
	sd := newSessionData(time.Hour)
	sd.values["foo"] = now
	ctx := s.addSessionDataToContext(context.Background(), sd)

	tm := s.PopTime(ctx, "foo")
	if tm != now {
		t.Errorf("got %v: expected %v", tm, now)
	}

	_, ok := sd.values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}

	tm = s.PopTime(ctx, "baz")
	if !tm.IsZero() {
		t.Errorf("got %v: expected %v", tm, time.Time{})
	}

}

func TestStatus(t *testing.T) {
	t.Parallel()

	s := New()
	sd := newSessionData(time.Hour)
	ctx := s.addSessionDataToContext(context.Background(), sd)

	status := s.Status(ctx)
	if status != Unmodified {
		t.Errorf("got %d: expected %d", status, Unmodified)
	}

	s.Put(ctx, "foo", "bar")
	status = s.Status(ctx)
	if status != Modified {
		t.Errorf("got %d: expected %d", status, Modified)
	}

	if err := s.Destroy(ctx); err != nil {
		t.Errorf("unexpected error destroying session data: %v", err)
	}

	status = s.Status(ctx)
	if status != Destroyed {
		t.Errorf("got %d: expected %d", status, Destroyed)
	}
}
