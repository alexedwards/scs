package firestore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFind(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ref := client.Collection("Sessions")
	_, err = ref.Doc("session_token").Set(ctx, map[string]interface{}{"Data": []byte("encoded_data"), "Expiry": time.Now().Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}

	m := NewWithCleanupInterval(client, 0)

	b, found, err := m.FindCtx(ctx, "session_token")
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

func TestFindMissing(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)

	_, found, err := m.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)

	err = m.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	ref := client.Collection("Sessions")
	ds, err := ref.Doc("session_token").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	d := ds.Data()
	if reflect.DeepEqual(d["Data"], []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", d["Data"], []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)
	ref := client.Collection("Sessions")

	_, err = ref.Doc("session_token").Set(ctx, map[string]interface{}{"Data": []byte("encoded_data"), "Expiry": time.Now().Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	err = m.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	ds, err := ref.Doc("session_token").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	d := ds.Data()
	if reflect.DeepEqual(d["Data"], []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", d["Data"], []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)

	err = m.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := m.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = m.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)
	ref := client.Collection("Sessions")

	_, err = ref.Doc("session_token").Set(ctx, map[string]interface{}{"Data": []byte("encoded_data"), "Expiry": time.Now().Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}

	err = m.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ref.Doc("session_token").Get(ctx)

	if err != nil {
		if status.Code(err) != codes.NotFound {
			t.Fatal(err)
		}
	} else {
		t.Fatal("session token not deleted")
	}
}

func TestAll(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)
	ref := client.Collection("Sessions")
	sessions := make(map[string][]byte)
	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("token_%v", i)
		val := []byte(key)
		_, err = ref.Doc(key).Set(ctx, map[string]interface{}{"Data": val, "Expiry": time.Now().Add(time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
		sessions[key] = val
	}

	gotSessions, err := m.AllCtx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for k := range sessions {
		err = m.DeleteCtx(ctx, k)
		if err != nil {
			t.Fatal(err)
		}
	}
	if reflect.DeepEqual(sessions, gotSessions) == false {
		t.Fatalf("got %v: expected %v", gotSessions, sessions)
	}
}

func TestCleanup(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 200*time.Millisecond)
	ref := client.Collection("Sessions")

	_, err = ref.Doc("session_token").Set(ctx, map[string]interface{}{"Data": []byte("encoded_data"), "Expiry": time.Now().Add(100 * time.Millisecond)})
	if err != nil {
		t.Fatal(err)
	}
	defer m.StopCleanup()

	_, err = ref.Doc("session_token").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(300 * time.Millisecond)
	_, err = ref.Doc("session_token").Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			t.Fatal(err)
		}
	} else {
		t.Fatal("session token not expired")
	}
}

func TestStopNilCleanup(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	m := NewWithCleanupInterval(client, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	m.StopCleanup()
}
