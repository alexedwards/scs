# boltstore

A [NATS JetStream KVStore](https://docs.nats.io/nats-concepts/jetstream/key-value-store) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [open a NATS KV store](https://natsbyexample.com/examples/kv/intro/go), and pass the database to `natsstore.New()` to establish the session store.

## Expiring contexts

You should probably be using the `CtxStore` methods for best context control. If, however, you decide to use the `Store` methods, you **must** set a global context timeout value.

```go
// set the global context timeout to 100ms
natsstore.New(db, WithTimeout(time.Millisecond * 100))
```

## Expired Session Cleanup

This package provides a background 'cleanup' goroutine to delete expired session data. This stops the database table from holding on to invalid sessions indefinitely and growing unnecessarily large. By default the cleanup runs every 1 minute. You can change this by using the `WithCleanupInterval` function to initialize your session store. For example:

```go
// Run a cleanup every 5 minutes.
natsstore.New(db, WithCleanupInterval(5*time.Minute))

// Disable the cleanup goroutine by setting the cleanup interval to zero.
natsstore.New(db, WithCleanupInterval(0))
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method.

## Notes

Currently Nats doesn't allow per-key expiry. In order to support per-key expiry, we take a rather hacky approach to including the expiry in the stored data that is checked on retrieval. Per-key expiry is in the works for release 2.11. Once this is available in the go client we will simplify the code and release as a /v2.
