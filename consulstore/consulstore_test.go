package consulstore

import (
	"bytes"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

func TestCommit(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 0, "scs:session:")
	c.kv.Delete(c.prefix+"key1", nil)

	c.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))

	pair, _, err := c.kv.Get(c.prefix+"key1", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pair.Value[8:], []byte("value1")) {
		t.Fatalf("expected bytes `value1`, got %s", pair.Value)
	}
}

func TestFind(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 0, "scs:session:")
	c.kv.Delete(c.prefix+"key1", nil)

	c.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	v, found, err := c.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if !bytes.Equal(v, []byte("value1")) {
		t.Fatalf("got %v: expected %v", v, []byte("value1"))
	}

	v, found, err = c.Find("key2")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if v != nil {
		t.Fatalf("got %v, expected %v", v, nil)
	}
}

func TestDelete(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 0, "scs:session:")
	c.kv.Delete(c.prefix+"key1", nil)

	err = c.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := c.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v, expected %v", found, true)
	}

	err = c.Delete("key1")
	if err != nil {
		t.Fatal(err)
	}

	_, found, err = c.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v, expected %v", found, false)
	}
}

func TestExpire(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 0, "scs:session:")
	c.kv.Delete(c.prefix+"session_token", nil)

	err = c.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := c.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)

	_, found, _ = c.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCleanup(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 2*time.Second, "scs:session:")
	c.kv.Delete(c.prefix+"session_token", nil)

	err = c.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	pair, _, err := c.kv.Get(c.prefix+"session_token", nil)
	if err != nil {
		t.Fatal(err)
	}

	if pair != nil {
		t.Fatalf("expected nil, got %v", pair)
	}

	c.StopCleanup()
}

func TestStopNilCleanup(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	c := NewWithOptions(cli, 0, "scs:session:")
	time.Sleep(100 * time.Millisecond)

	// A send to a nil channel will block forever
	c.StopCleanup()
}
