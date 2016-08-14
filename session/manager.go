package session

import (
	"bytes"
	"net/http"

	"github.com/alexedwards/scs"
)

type Middleware func(h http.Handler) http.Handler

func Manage(engine scs.Engine, opts ...Option) Middleware {
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
	engine scs.Engine
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

func defaultErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
