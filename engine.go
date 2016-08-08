package scs

import "time"

type Engine interface {
	Delete(token string) (err error)
	FindValues(token string) (b []byte, found bool, err error)
	Save(token string, b []byte, maxAge time.Duration) (err error)
}
