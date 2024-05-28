package natsstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var db jetstream.KeyValue

var natsURL string

func localNats() *server.Server {
	ns, err := server.NewServer(&server.Options{JetStream: true})
	if err != nil {
		panic(err)
	}

	go ns.Start()
	time.Sleep(time.Second)
	return ns
}

func TestMain(m *testing.M) {
	ns := localNats()
	defer ns.Shutdown()

	natsURL = ns.ClientURL()

	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  "scs_tests",
		Storage: jetstream.MemoryStorage,
	})
	if err != nil {
		panic(err)
	}

	results := m.Run()

	os.Exit(results)
}

type testsData struct {
	key    string
	value  []byte
	expiry time.Time
}

func generateData(count int, expiry time.Duration) []testsData {
	val := func() []byte {
		out := make([]byte, 8)
		rand.Read(out)
		return out
	}
	// mimics what scs uses to generate tokens
	key := func() string {
		return base64.RawURLEncoding.EncodeToString(val())
	}
	out := make([]testsData, count)
	for i := 0; i < count; i++ {
		out[i] = testsData{key(), val(), time.Now().Add(expiry)}
	}
	return out
}

func TestCrud(t *testing.T) {
	h := New(db, WithTimeout(time.Minute))

	src := generateData(1, time.Hour)[0]

	t.Run("Commit", func(t *testing.T) {
		err := h.Commit(src.key, src.value, src.expiry)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Find", func(t *testing.T) {
		val, found, err := h.Find(src.key)
		if err != nil {
			t.Error(err)
		}

		if found != true {
			t.Error("record not found")
		}

		if !bytes.Equal(val, src.value) {
			t.Error("values don't match")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := h.Delete(src.key)
		if err != nil {
			t.Error(err)
		}

		_, found, err := h.Find(src.key)
		if err != nil {
			t.Error(err)
		}

		if found != false {
			t.Error("record not deleted")
		}
	})
}

func TestCrudCtx(t *testing.T) {
	h := New(db)

	src := generateData(1, time.Hour)[0]

	t.Run("Commit", func(t *testing.T) {
		err := h.CommitCtx(context.TODO(), src.key, src.value, src.expiry)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Find", func(t *testing.T) {
		val, found, err := h.FindCtx(context.TODO(), src.key)
		if err != nil {
			t.Error(err)
		}

		if found != true {
			t.Error("record not found")
		}

		if !bytes.Equal(val, src.value) {
			t.Error("values don't match")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		ctx := context.TODO()
		err := h.DeleteCtx(ctx, src.key)
		if err != nil {
			t.Error(err)
		}

		_, found, err := h.FindCtx(ctx, src.key)
		if err != nil {
			t.Error(err)
		}

		if found != false {
			t.Error("record not deleted")
		}
	})
}

func TestAll(t *testing.T) {
	h := New(db, WithTimeout(time.Second))
	src := generateData(2, time.Hour)

	for _, row := range src {
		h.Commit(row.key, row.value, row.expiry)
	}

	t.Run("basic", func(t *testing.T) {
		all, err := h.All()
		if err != nil {
			t.Error(err)
		}

		if len(all) != len(src) {
			t.Error("count of All is incorrect")
		}
	})
	t.Run("with deletes", func(t *testing.T) {
		h.Delete(src[0].key)

		all, err := h.All()
		if err != nil {
			t.Error(err)
		}

		if len(all) != len(src)-1 {
			t.Error("count of All is incorrect")
		}
	})
	t.Run("with expiry", func(t *testing.T) {
		before, _ := h.All()

		shorties := generateData(2, time.Second*2)
		for _, row := range shorties {
			h.Commit(row.key, row.value, row.expiry)
		}

		after, _ := h.All()
		if len(after) != len(before)+len(shorties) {
			t.Error("unexpected lengths from All")
		}

		time.Sleep(time.Second * 3)

		after, _ = h.All()
		if len(after) != len(before) {
			t.Error("unexpected lengths from All after expiry")
		}
	})

	// cleanup
	all, _ := h.All()
	for key := range all {
		h.Delete(key)
	}
}

func TestCleanup(t *testing.T) {
	h := New(db, WithTimeout(time.Second), WithCleanupInterval(time.Second))
	src := generateData(2, time.Hour)

	// uneccessary
	defer h.StopCleanup()

	for _, row := range src {
		h.Commit(row.key, row.value, row.expiry)
	}

	time.Sleep(2 * time.Second)

	for _, row := range src {
		h.Delete(row.key)
	}

	time.Sleep(2 * time.Second)
}
