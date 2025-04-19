package mongodbstore

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type item struct {
	Token      string `json:"token"`
	Object     []byte `json:"object"`
	Expiration int64  `json:"expiration"`
}

// MongoDBStore represents the session store.
type MongoDBStore struct {
	collection  *mongo.Collection
	stopCleanup chan bool
}

type Config struct {
	// CleanUpInterval is the interval between each cleanup operation.
	// If set to 0, the cleanup operation is disabled.
	CleanupInterval time.Duration

	// CollectionName is the name of the collection where the session data will be stored.
	// If not set, it will default to "sessions".
	CollectionName string
}

// New returns a new MongoDBStore instance, with a background cleanup goroutine that
// runs every minute to remove expired session data.
func New(db *mongo.Database) *MongoDBStore {
	return NewWithConfig(db, Config{
		CleanupInterval: time.Minute,
	})
}

// NewWithCleanupInterval returns a new MongoDBStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(db *mongo.Database, cleanupInterval time.Duration) *MongoDBStore {
	return NewWithConfig(db, Config{
		CleanupInterval: cleanupInterval,
	})
}

// NewWithConfig returns a new MongoDBStore instance with the given configuration.
// If the CollectionName field is empty, it will be set to "sessions".
// If the CleanupInterval field is 0, the cleanup goroutine will not be started.
func NewWithConfig(db *mongo.Database, config Config) *MongoDBStore {
	if config.CollectionName == "" {
		config.CollectionName = "sessions"
	}
	collection := db.Collection(config.CollectionName)

	m := &MongoDBStore{
		collection: collection,
	}

	if config.CleanupInterval > 0 {
		go m.startCleanup(config.CleanupInterval)
	}

	return m
}

// Find returns the data for a given session token from the MongoDBStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (m *MongoDBStore) Find(token string) ([]byte, bool, error) {
	filter := bson.M{"token": token}
	result := m.collection.FindOne(context.Background(), filter)

	err := result.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false, nil
		}

		return nil, false, err
	}

	var i item
	err = result.Decode(&i)
	if err != nil {
		return nil, false, err
	}

	if time.Now().UnixNano() > i.Expiration {
		return nil, false, nil
	}

	return i.Object, true, nil
}

// Commit adds a session token and data to the MongoDBStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (m *MongoDBStore) Commit(token string, b []byte, expiry time.Time) error {
	item := item{
		Token:      token,
		Object:     b,
		Expiration: expiry.UnixNano(),
	}

	// Create or replace the existing item
	// https://docs.mongodb.com/drivers/node/fundamentals/crud/write-operations/upsert/
	filter := bson.M{"token": token}
	update := bson.M{"$set": item}
	opts := options.Update().SetUpsert(true)
	_, err := m.collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

// Delete removes a session token and corresponding data from the MongoDBStore
// instance.
func (m *MongoDBStore) Delete(token string) error {
	filter := bson.M{"token": token}
	_, err := m.collection.DeleteOne(context.Background(), filter)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the MongoDBStore instance.
func (m *MongoDBStore) All() (map[string][]byte, error) {
	filter := bson.D{}
	cursor, err := m.collection.Find(context.Background(), filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	sessions := make(map[string][]byte)

	for cursor.Next(context.Background()) {
		var i item

		err := cursor.Decode(&i)
		if err != nil {
			return nil, err
		}

		if i.Expiration > time.Now().UnixNano() {
			sessions[i.Token] = i.Object
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	cursor.Close(context.Background())

	return sessions, nil
}

func (m *MongoDBStore) startCleanup(cleanupInterval time.Duration) {
	m.stopCleanup = make(chan bool)
	ticker := time.NewTicker(cleanupInterval)
	for {
		select {
		case <-ticker.C:
			err := m.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-m.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the MongoDBStore
// instance. It's rare to terminate this; generally MongoDBStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the MongoDBStore is transient.
// An example is creating a new MongoDBStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// MongoDBStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (m *MongoDBStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *MongoDBStore) deleteExpired() error {
	now := time.Now().UnixNano()
	filter := bson.M{"expiration": bson.M{"$lt": now}}
	_, err := m.collection.DeleteMany(context.Background(), filter, nil)
	if err != nil {
		return err
	}

	return nil
}
