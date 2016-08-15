package session

import "time"

type Engine interface {
	Delete(token string) (err error)
	Find(token string) (b []byte, found bool, err error)
	Save(token string, b []byte, expiry time.Time) (err error)
}
