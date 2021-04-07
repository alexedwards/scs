package mongodbstore

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestFind(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	filter := bson.M{"token": "session_token"}
	update := bson.M{"$set": item{Token: "session_token", Object: []byte("encoded_data"), Expiration: time.Now().Add(time.Second).UnixNano()}}
	opts := options.Update().SetUpsert(true)
	_, err = m.collection.UpdateOne(context.Background(), filter, update, opts)

	b, found, err := m.Find("session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	_, found, err := m.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestCommitNew(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	err = m.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	filter := bson.M{"token": "session_token"}
	result := m.collection.FindOne(context.Background(), filter)

	err = result.Err()
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	var i item
	err = result.Decode(&i)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	if reflect.DeepEqual(i.Object, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", i.Object, []byte("encoded_data"))
	}
}

func TestCommitUpdated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	err = m.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	err = m.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	filter := bson.M{"token": "session_token"}
	result := m.collection.FindOne(context.Background(), filter)

	err = result.Err()
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	var i item
	err = result.Decode(&i)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	if reflect.DeepEqual(i.Object, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", i.Object, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	err = m.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	_, found, _ := m.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(101 * time.Millisecond)
	_, found, _ = m.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 0)

	filter := bson.M{"token": "session_token"}
	update := bson.M{"$set": item{Token: "session_token", Object: []byte("encoded_data"), Expiration: time.Now().Add(time.Second).UnixNano()}}
	opts := options.Update().SetUpsert(true)
	_, err = m.collection.UpdateOne(context.Background(), filter, update, opts)

	err = m.Delete("session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	result := m.collection.FindOne(context.Background(), filter)

	err = result.Err()
	if err != mongo.ErrNoDocuments {
		t.Fatalf("got %v: expected %v", nil, mongo.ErrNoDocuments)
	}
}

func TestCleanupInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)

	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	m := NewWithCleanupInterval(client.Database("database"), 100*time.Millisecond)
	defer m.StopCleanup()

	filter := bson.M{"token": "session_token"}
	update := bson.M{"$set": item{Token: "session_token", Object: []byte("encoded_data"), Expiration: time.Now().Add(500 * time.Millisecond).UnixNano()}}
	opts := options.Update().SetUpsert(true)
	_, err = m.collection.UpdateOne(context.Background(), filter, update, opts)

	result := m.collection.FindOne(context.Background(), filter)

	err = result.Err()
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	var i item
	err = result.Decode(&i)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}

	time.Sleep(time.Second)
	result = m.collection.FindOne(context.Background(), filter)

	err = result.Err()
	if err != mongo.ErrNoDocuments {
		t.Fatalf("got %v: expected %v", nil, mongo.ErrNoDocuments)
	}
}
