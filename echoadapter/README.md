# echoadapter

A [Echo](https://github.com/labstack/echo) based session adapter for [SCS](https://github.com/alexedwards/scs).

## Example

```go
package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/echoadapter"
	"github.com/labstack/echo/v4"
)

var sessionManager *scs.SessionManager

func main() {
	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	// Initialize a new Echo instance.
	e := echo.New()

	// Wrap your handlers with the LoadAndSave() adapter middleware.
	sessionAdapter := echoadapter.New(sessionManager)
	e.Use(sessionAdapter.LoadAndSave)

	e.GET("/put", putHandler)
	e.GET("/get", getHandler)

	e.Logger.Fatal(e.Start(":4000"))
}

func putHandler(c echo.Context) error {
	// Store a new key and value in the session data.
	sessionManager.Put(c.Request().Context(), "message", "Hello from a Echo session!")
	return nil
}

func getHandler(c echo.Context) error {
	// Use the GetString helper to retrieve the string value associated with a
	// key. The zero value is returned if the key does not exist.
	msg := sessionManager.GetString(c.Request().Context(), "message")
	return c.String(http.StatusOK, msg)
}
```