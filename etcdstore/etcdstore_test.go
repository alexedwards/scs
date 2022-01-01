package etcdstore

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestFind(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.client.Put(ctx, e.prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := e.FindCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestSaveNew(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	err = e.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	res, err := e.client.Get(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Kvs) == 0 {
		t.Fatalf("missing key")
	}

	if reflect.DeepEqual(res.Kvs[0].Value, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", res.Kvs[0].Value, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"missing_session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := e.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveUpdated(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.client.Put(ctx, e.prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	err = e.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	res, err := e.client.Get(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Kvs) == 0 {
		t.Fatalf("missing key")
	}

	if reflect.DeepEqual(res.Kvs[0].Value, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", res.Kvs[0].Value, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	err = e.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := e.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(3 * time.Second)

	_, found, _ = e.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	e := New(cli)
	ctx := context.Background()

	_, err = e.client.Delete(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.client.Put(ctx, e.prefix+"session_token", "encoded_data")
	if err != nil {
		t.Fatal(err)
	}

	err = e.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	res, err := e.client.Get(ctx, e.prefix+"session_token")
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Kvs) != 0 {
		t.Fatalf("got %v: expected %v", res.Kvs[0].Value, nil)
	}
}
