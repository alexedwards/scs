//go:build go1.18
// +build go1.18

package scs

import (
	"bufio"
	"net"
	"net/http"
)

func (sw *sessionResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return sw.ResponseWriter.(http.Hijacker).Hijack()
}
