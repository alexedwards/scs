package fiberadapter

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gofiber/fiber/v2"
)

// FiberAdapter represents the session adapter.
type FiberAdapter struct {
	*scs.SessionManager
}

// New returns a new FiberAdapter instance that embeds the original SCS session manager.
func New(s *scs.SessionManager) *FiberAdapter {
	return &FiberAdapter{s}
}

// LoadAndSave provides a Fiber middleware which automatically loads and saves session
// data for the current request, and communicates the session token to and from
// the client in a cookie.
func (a *FiberAdapter) LoadAndSave() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(a.Cookie.Name)

		ctx, err := a.Load(c.UserContext(), token)
		if err != nil {
			return err
		}

		c.SetUserContext(ctx)
		c.Next()

		if form, err := c.MultipartForm(); err == nil {
			form.RemoveAll()
		}

		switch a.Status(ctx) {
		case scs.Modified:
			token, expiry, err := a.Commit(ctx)
			if err != nil {
				return err
			}

			a.writeSessionCookie(c, token, expiry)
		case scs.Destroyed:
			a.writeSessionCookie(c, "", time.Time{})
		}

		c.Vary("Cookie")

		return nil
	}
}

func (a *FiberAdapter) writeSessionCookie(c *fiber.Ctx, token string, expiry time.Time) {
	cookie := &http.Cookie{
		Name:     a.Cookie.Name,
		Value:    token,
		Path:     a.Cookie.Path,
		Domain:   a.Cookie.Domain,
		Secure:   a.Cookie.Secure,
		HttpOnly: a.Cookie.HttpOnly,
		SameSite: a.Cookie.SameSite,
	}

	if expiry.IsZero() {
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else if a.Cookie.Persist || a.GetBool(c.UserContext(), "__rememberMe") {
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)        // Round up to the nearest second.
		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1) // Round up to the nearest second.
	}

	c.Set("Set-Cookie", cookie.String())
	c.Set("Cache-Control", `no-cache="Set-Cookie"`)
}
