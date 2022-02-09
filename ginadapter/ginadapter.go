package ginadapter

import (
	"bytes"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

// GinAdapter represents the session adapter.
type GinAdapter struct {
	*scs.SessionManager
}

// New returns a new GinAdapter instance that embeds the original SCS session manager.
func New(s *scs.SessionManager) *GinAdapter {
	return &GinAdapter{s}
}

// LoadAndSave provides a Gin middleware which automatically loads and saves session
// data for the current request, and communicates the session token to and from
// the client in a cookie.
func (a *GinAdapter) LoadAndSave(c *gin.Context) {
	w := c.Writer
	r := c.Request

	var token string
	cookie, err := r.Cookie(a.Cookie.Name)
	if err == nil {
		token = cookie.Value
	}

	ctx, err := a.Load(r.Context(), token)
	if err != nil {
		a.ErrorFunc(w, r, err)
		return
	}

	sr := r.WithContext(ctx)
	bw := &bufferedResponseWriter{ResponseWriter: w}

	c.Request = sr
	c.Next()

	if sr.MultipartForm != nil {
		sr.MultipartForm.RemoveAll()
	}

	switch a.Status(ctx) {
	case scs.Modified:
		token, expiry, err := a.Commit(ctx)
		if err != nil {
			a.ErrorFunc(w, r, err)
			return
		}

		a.WriteSessionCookie(ctx, w, token, expiry)
	case scs.Destroyed:
		a.WriteSessionCookie(ctx, w, "", time.Time{})
	}

	w.Header().Add("Vary", "Cookie")

	if bw.code != 0 {
		w.WriteHeader(bw.code)
	}
	w.Write(bw.buf.Bytes())
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	code        int
	wroteHeader bool
}

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	return bw.buf.Write(b)
}

func (bw *bufferedResponseWriter) WriteHeader(code int) {
	if !bw.wroteHeader {
		bw.code = code
		bw.wroteHeader = true
	}
}
