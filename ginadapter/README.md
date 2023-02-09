# ginadapter

A [Gin](https://github.com/gin-gonic/gin) based session adapter for [SCS](https://github.com/alexedwards/scs).

## Example

```go
package main

import (
	"io"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/ginadapter"
	"github.com/gin-gonic/gin"
)

var sessionManager *scs.SessionManager

func main() {
	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	// Initialize a new Gin instance.
	r := gin.Default()

	// Wrap your handlers with the LoadAndSave() adapter middleware.
	sessionAdapter := ginadapter.New(sessionManager)
	r.Use(sessionAdapter.LoadAndSave)

	r.GET("/put", putHandler)
	r.GET("/get", getHandler)

	r.Run(":4000")
}

func putHandler(c *gin.Context) {
	// Store a new key and value in the session data.
	sessionManager.Put(c.Request.Context(), "message", "Hello from a Gin session!")
}

func getHandler(c *gin.Context) {
	// Use the GetString helper to retrieve the string value associated with a
	// key. The zero value is returned if the key does not exist.
	msg := sessionManager.GetString(c.Request.Context(), "message")
	io.WriteString(c.Writer, msg)
}
```