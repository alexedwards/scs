package session

import (
	"bytes"
	"net/http"
)

// Deprecated: Middleware previously defined the signature for the session management
// middleware returned by Manage. Manage now returns a func(h http.Handler) http.Handler
// directly instead, so it's easier to use with middleware chaining packages like Alice.
type Middleware func(h http.Handler) http.Handler

/*
Manage returns a new session manager middleware instance. The first parameter
should be a valid storage engine, followed by zero or more functional options.

For example:

	session.Manage(memstore.New(0))

	session.Manage(memstore.New(0), session.Lifetime(14*24*time.Hour))

	session.Manage(memstore.New(0),
		session.Secure(true),
		session.Persist(true),
		session.Lifetime(14*24*time.Hour),
	)

The returned session manager can be used to wrap any http.Handler. It automatically
loads sessions based on the HTTP request and saves session data as and when necessary.
*/
func Manage(engine Engine, opts ...Option) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		do := *defaultOptions

		m := &manager{
			h:      h,
			engine: engine,
			opts:   &do,
		}

		for _, option := range opts {
			option(m.opts)
		}

		return m
	}
}

type manager struct {
	h      http.Handler
	engine Engine
	opts   *options
}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sr, err := load(r, m.engine, m.opts)
	if err != nil {
		m.opts.errorFunc(w, r, err)
		return
	}
	bw := &bufferedResponseWriter{ResponseWriter: w}
	m.h.ServeHTTP(bw, sr)

	err = write(w, sr)
	if err != nil {
		m.opts.errorFunc(w, r, err)
		return
	}

	if bw.code != 0 {
		w.WriteHeader(bw.code)
	}
	w.Write(bw.buf.Bytes())
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf  bytes.Buffer
	code int
}

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	return bw.buf.Write(b)
}

func (bw *bufferedResponseWriter) WriteHeader(code int) {
	bw.code = code
}

func (bw *bufferedResponseWriter) Flush() {
	f, ok := bw.ResponseWriter.(http.Flusher)
	if ok == true {
		bw.ResponseWriter.Write(bw.buf.Bytes())
		f.Flush()
		bw.buf.Reset()
	}
}

func defaultErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
