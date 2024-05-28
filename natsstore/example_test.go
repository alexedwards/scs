package natsstore

import (
	"context"
	"fmt"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func ExampleRun() {
	nc, _ := nats.Connect(natsURL)
	defer nc.Drain()

	js, _ := jetstream.New(nc)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, _ := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "scs_example",
		Storage: jetstream.MemoryStorage,
	})

	store := New(db, WithTimeout(time.Second), WithCleanupInterval(30*time.Minute))

	sessionManager := scs.New()
	sessionManager.Store = store

	// see the store in action

	putCtx, _ := sessionManager.Load(context.Background(), "")

	sessionManager.Put(putCtx, "foo", "bar")
	token, _, _ := sessionManager.Commit(putCtx)

	getCtx, _ := sessionManager.Load(context.Background(), token)

	foo := sessionManager.GetString(getCtx, "foo")

	fmt.Println(foo)
	// Output: bar
}
