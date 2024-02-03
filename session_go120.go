//go:build go1.20
// +build go1.20

package scs

import (
	"bufio"
	"net"
	"net/http"
)

func (sw *sessionResponseWriter) Flush() {
	http.NewResponseController(sw.ResponseWriter).Flush()
}

func (sw *sessionResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(sw.ResponseWriter).Hijack()
}
