# consulstore

A [Consul](https://github.com/hashicorp/consul) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [setup a connection](https://github.com/hashicorp/consul/tree/main/api#usage), and pass the connection to `consulstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/consulstore"
	"github.com/hashicorp/consul/api"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to Consul.
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a new session manager and configure it to use consulstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = consulstore.New(cli)

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

Consul uses [sessions](https://www.consul.io/api/session#ttl) to apply TTL's to keys in the K/V store with a max-lifetime of 24hrs.
To avoid complexities and restrictions this package provides a background 'cleanup' goroutine to delete expired session data. This stops the store from holding on to invalid sessions indefinitely and growing unnecessarily large. By default the cleanup runs every 1 minute. You can change this by using the `NewWithOptions()` function to initialize your session store. For example:

```go
// Run a cleanup every 5 minutes.
consulstore.NewWithOptions(db, 5*time.Minute, "scs:session:")

// Disable the cleanup goroutine by setting the cleanup interval to zero.
consulstore.NewWithOptions(db, 0, "scs:session:")
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method. For example:

```go
func TestExample(t *testing.T) {
	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	store := consulstore.New(cli)
	defer store.StopCleanup()

	sessionManager = scs.New()
	sessionManager.Store = store

	// Run test...
}
```

## Key Collisions

By default keys are in the form `scs:session:<token>`. For example:

```
"scs:session:ZnirGwi2FiLwXeVlP5nD77IpfJZMVr6un9oZu2qtJrg"
```

Because the token is highly unique, key collisions are not a concern. But if you're configuring *multiple session managers*, both of which use `consulstore`, then you may want the keys to have a different prefix depending on which session manager wrote them. You can do this by using the `NewWithOptions()` method like so:

```go
cli, err := api.NewClient(api.DefaultConfig())
if err != nil {
	log.Fatal(err)
}

sessionManagerOne = scs.New()
sessionManagerOne.Store = consulstore.NewWithOptions(cli, time.Minute, "scs:session:1:")

sessionManagerTwo = scs.New()
sessionManagerTwo.Store = consulstore.NewWithOptions(cli, time.Minute, "scs:session:2:")
```
