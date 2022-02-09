# fiberadapter

A [Fiber](https://github.com/gofiber/fiber) based session adapter for [SCS](https://github.com/alexedwards/scs).

## Example

```go
package main

import (
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/fiberadapter"
	"github.com/gofiber/fiber/v2"
)

var sessionManager *scs.SessionManager

func main() {
	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	// Initialize a new Fiber instance.
	app := fiber.New()

	// Wrap your handlers with the LoadAndSave() adapter middleware.
	sessionAdapter := fiberadapter.New(sessionManager)
	app.Use(sessionAdapter.LoadAndSave())

	app.Get("/put", putHandler)
	app.Get("/get", getHandler)

	app.Listen(":4000")
}

func putHandler(c *fiber.Ctx) error {
	// Store a new key and value in the session data.
	sessionManager.Put(c.UserContext(), "message", "Hello from a Fiber session!")
	return nil
}

func getHandler(c *fiber.Ctx) error {
	// Use the GetString helper to retrieve the string value associated with a
	// key. The zero value is returned if the key does not exist.
	msg := sessionManager.GetString(c.UserContext(), "message")
	return c.SendString(msg)
}
```