package firestore

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FireStore represents the session store.
type FireStore struct {
	*firestore.Client
	Sessions    *firestore.CollectionRef
	stopCleanup chan bool
}

type sessionDoc struct {
	Data   []byte
	Expiry time.Time
}

// New returns a new FireStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func New(client *firestore.Client) *FireStore {
	return NewWithCleanupInterval(client, 5*time.Minute)
}

// NewWithCleanupInterval returns a new FireStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(client *firestore.Client, cleanupInterval time.Duration) *FireStore {
	m := &FireStore{
		Client:   client,
		Sessions: client.Collection("Sessions"),
	}

	if cleanupInterval > 0 {
		go m.startCleanup(cleanupInterval)
	}

	return m
}

// FindCtx returns the data for a given session token from the FireStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (m *FireStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	ds, err := m.Sessions.Doc(token).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	var sd sessionDoc
	err = ds.DataTo(&sd)
	if err != nil {
		return nil, false, err
	}
	if time.Now().After(sd.Expiry) {
		return nil, false, nil
	}
	return sd.Data, true, nil
}

// CommitCtx adds a session token and data to the FireStore instance with the given
// expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (m *FireStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	sd := sessionDoc{Data: b, Expiry: expiry}
	_, err := m.Sessions.Doc(token).Set(ctx, &sd)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCtx removes a session token and corresponding data from the FireStore
// instance.
func (m *FireStore) DeleteCtx(ctx context.Context, token string) error {
	_, err := m.Sessions.Doc(token).Delete(ctx)
	return err
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the firestore instance.
func (m *FireStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	iter := m.Sessions.Where("Expiry", ">=", time.Now()).Documents(ctx)
	defer iter.Stop()
	sessions := make(map[string][]byte)
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var sd sessionDoc
		err = snap.DataTo(&sd)
		if err != nil {
			return nil, err
		}
		sessions[snap.Ref.ID] = sd.Data
	}
	return sessions, nil
}

func (m *FireStore) startCleanup(interval time.Duration) {
	m.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
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

// StopCleanup terminates the background cleanup goroutine for the FireStore
// instance. It's rare to terminate this; generally FireStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the FireStore is transient.
// An example is creating a new FireStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// MySQLStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (m *FireStore) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *FireStore) deleteExpired() error {
	ctx := context.Background()
	iter := m.Sessions.Where("Expiry", "<", time.Now()).Documents(ctx)
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate: %v", err)
			break
		}
		_, err = snap.Ref.Delete(ctx, firestore.LastUpdateTime(snap.UpdateTime))
		if err != nil {
			log.Printf("Failed to delete: %v", err)
			continue
		}
	}
	iter.Stop()
	return nil
}

// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a seperate type makes any errors caught
// only at runtime instead of compile time. Oh well.

func (m *FireStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}
func (m *FireStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}
func (m *FireStore) Delete(token string) error {
	panic("missing context arg")
}
