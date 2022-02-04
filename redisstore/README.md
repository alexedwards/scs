# redisstore

A [Redis](https://github.com/gomodule/redigo) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [setup a connection pool](https://godoc.org/github.com/gomodule/redigo/redis#Pool), and pass the pool to `redisstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/redisstore"
	"github.com/gomodule/redigo/redis"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection pool to Redis.
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "host:6379")
		},
	}
	
	// Initialize a new session manager and configure it to use redisstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = redisstore.New(pool)

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

Because the token is highly unique, key collisions are not a concern. But if you're configuring *multiple session managers*, both of which use `redisstore`, then you may want the keys to have a different prefix depending on which session manager wrote them. You can do this by using the `NewWithPrefix()` method like so:

```go
pool := &redis.Pool{
    MaxIdle: 10,
    Dial: func() (redis.Conn, error) {
        return redis.Dial("tcp", "host:6379")
    },
}

sessionManagerOne = scs.New()
sessionManagerOne.Store = redisstore.NewWithPrefix(pool, "scs:session:1:")

sessionManagerTwo = scs.New()
sessionManagerTwo.Store = redisstore.NewWithPrefix(pool, "scs:session:2:")
```
