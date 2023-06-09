package coherencestore

import (
	"context"
	"github.com/oracle/coherence-go-client/coherence"
	"log"
	"testing"
	"time"
)

// TestCoherenceWithContext tests the functions that pass Context.
func TestCoherenceWithContext(t *testing.T) {
	ctx := context.Background()
	// connect to default of localhost:1408
	session, err := coherence.NewSession(ctx, coherence.WithPlainText())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	c, err := New(session)
	if err != nil {
		t.Fatal(err)
	}

	err = c.CommitCtx(ctx, "key1", []byte("value1"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	data, found, err := c.FindCtx(ctx, "key1")
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		t.Fatal("key1 should have been found")
	}

	t.Logf("Value of key1 is %v", string(data))

	t.Logf("Commit key2")
	err = c.CommitCtx(ctx, "key2", []byte("value2"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Commit key3")
	err = c.CommitCtx(ctx, "key3", []byte("value3"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Get All")
	values, err := c.AllCtx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("number of values is %v", len(values))
	for k, v := range values {
		t.Logf("key=%v, value=%v", k, string(v))
	}

	t.Logf("Delete key2")
	err = c.DeleteCtx(ctx, "key2")
	if err != nil {
		t.Fatal(err)
	}

	data, found, err = c.FindCtx(ctx, "key2")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("key2 found=%v", found)

	// sleep for 11 seconds and the other two entries should have gone
	t.Logf("Sleeping for 11 seconds")
	time.Sleep(time.Duration(21) * time.Second)

	t.Logf("Get All after sleep")
	values, err = c.AllCtx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Number of entries=%v", len(values))
}

// TestCoherenceWithOutContext tests functions that do not pass Context.
func TestCoherenceWithOutContext(t *testing.T) {
	ctx := context.Background()
	// connect to default of localhost:1408
	session, err := coherence.NewSession(ctx, coherence.WithPlainText())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	c, err := New(session)
	if err != nil {
		t.Fatal(err)
	}

	testWithoutContext(t, c)
}

// TestCoherenceCustomCacheName tests functions that do not pass Context and use a custom session cache name.
func TestCoherenceCustomCacheName(t *testing.T) {
	ctx := context.Background()
	// connect to default of localhost:1408
	session, err := coherence.NewSession(ctx, coherence.WithPlainText())
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	c, err := NewWithCache(session, "my-session-cache")
	if err != nil {
		t.Fatal(err)
	}
	testWithoutContext(t, c)
}

func testWithoutContext(t *testing.T, c *CoherenceStore) {
	err := c.Commit("key1", []byte("value1"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	data, found, err := c.Find("key1")
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		t.Fatal("key1 should have been found")
	}

	t.Logf("Value of key1 is %v", string(data))

	t.Logf("Commit key2")
	err = c.Commit("key2", []byte("value2"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Commit key3")
	err = c.Commit("key3", []byte("value3"), time.Now().Add(time.Duration(10)*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Get All")
	values, err := c.All()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("number of values is %v", len(values))
	for k, v := range values {
		t.Logf("key=%v, value=%v", k, string(v))
	}

	t.Logf("Delete key2")
	err = c.Delete("key2")
	if err != nil {
		t.Fatal(err)
	}

	data, found, err = c.Find("key2")
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("key2 found=%v", found)

	// sleep for 11 seconds and the other two entries should have gone
	t.Logf("Sleeping for 11 seconds")
	time.Sleep(time.Duration(21) * time.Second)

	t.Logf("Get All after sleep")
	values, err = c.All()
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("Number of entries=%v", len(values))
}
