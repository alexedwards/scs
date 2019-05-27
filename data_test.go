package scs

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSessionDataFromContext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	s := NewSession()
	s.getSessionDataFromContext(context.Background())
}

func TestPut(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	ctx := s.addSessionDataToContext(context.Background(), sd)

	s.Put(ctx, "foo", "bar")

	if sd.Values["foo"] != "bar" {
		t.Errorf("got %q: expected %q", sd.Values["foo"], "bar")
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestGet(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str, ok := s.Pop(ctx, "foo").(string)
	if !ok {
		t.Errorf("could not convert %T to string", s.Get(ctx, "foo"))
	}

	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}

	_, ok = sd.Values["foo"]
	if ok {
		t.Errorf("got %v: expected %v", ok, false)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestRemove(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	s.Remove(ctx, "foo")

	if sd.Values["foo"] != nil {
		t.Errorf("got %v: expected %v", sd.Values["foo"], nil)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestClear(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	sd.Values["baz"] = "boz"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	s.Clear(ctx)

	if sd.Values["foo"] != nil {
		t.Errorf("got %v: expected %v", sd.Values["foo"], nil)
	}

	if sd.Values["baz"] != nil {
		t.Errorf("got %v: expected %v", sd.Values["baz"], nil)
	}

	if sd.status != Modified {
		t.Errorf("got %v: expected %v", sd.status, "modified")
	}
}

func TestExists(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	if !s.Exists(ctx, "foo") {
		t.Errorf("got %v: expected %v", s.Exists(ctx, "foo"), true)
	}

	if s.Exists(ctx, "baz") {
		t.Errorf("got %v: expected %v", s.Exists(ctx, "baz"), false)
	}
}

func TestKeys(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	sd.Values["woo"] = "waa"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	keys := s.Keys(ctx)
	if !reflect.DeepEqual(keys, []string{"foo", "woo"}) {
		t.Errorf("got %v: expected %v", keys, []string{"foo", "woo"})
	}
}

func TestGetString(t *testing.T) {
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = true
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = 123
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = 123.456
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = []byte("bar")
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
	now := time.Now()

	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = now
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
	s := NewSession()
	sd := newSessionData(time.Hour)
	sd.Values["foo"] = "bar"
	ctx := s.addSessionDataToContext(context.Background(), sd)

	str := s.PopString(ctx, "foo")
	if str != "bar" {
		t.Errorf("got %q: expected %q", str, "bar")
	}

	_, ok := sd.Values["foo"]
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

func TestStatus(t *testing.T) {
	s := NewSession()
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

	s.Destroy(ctx)
	status = s.Status(ctx)
	if status != Destroyed {
		t.Errorf("got %d: expected %d", status, Destroyed)
	}
}
