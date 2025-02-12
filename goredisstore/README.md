# goredisstore

A [Redis](github.com/redis/go-redis/v9) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [setup a client](https://pkg.go.dev/github.com/redis/go-redis/v9#NewClient), and pass the client to `goredisstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/goredisstore"
	"github.com/redis/go-redis/v9"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection pool to Redis.
	opt, err := redis.ParseURL("redis://localhost:6379")
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	defer client.Close()

	// Initialize a new session manager and configure it to use goredisstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = goredisstore.New(client)

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	http.ListenAndServe(":4000", sessionManager.LoadAndSave(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "message", "Hello from a session!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	msg := sessionManager.GetString(r.Context(), "message")
	io.WriteString(w, msg)
}
```

## Expired Session Cleanup

Redis will [automatically remove](http://redis.io/commands/expire#how-redis-expires-keys) expired session keys.

## Key Collisions

By default keys are in the form `scs:session:<token>`. For example:

```
"scs:session:ZnirGwi2FiLwXeVlP5nD77IpfJZMVr6un9oZu2qtJrg"
```

Because the token is highly unique, key collisions are not a concern. But if you're configuring *multiple session managers*, both of which use `goredisstore`, then you may want the keys to have a different prefix depending on which session manager wrote them. You can do this by using the `NewWithPrefix()` method like so:

```go
opt, err := redis.ParseURL("redis://localhost:6379")
if err != nil {
    panic(err)
}
client := redis.NewClient(opt)

sessionManagerOne = scs.New()
sessionManagerOne.Store = goredisstore.NewWithPrefix(client, "scs:session:1:")

sessionManagerTwo = scs.New()
sessionManagerTwo.Store = goredisstore.NewWithPrefix(client, "scs:session:2:")
```
## Iterating over all Sessions

If you intend to use the sessionstore.Iterate() function to iterate over all
sessions on a busy Redis server with many keys stored, be warned that this
can take a long time and is therefore probably only interesting for debugging
purposes.
