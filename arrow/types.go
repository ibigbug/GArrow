package arrow

import "errors"

type Runnable interface {
	Run() error
}

var (
	// ErrReused is not an error, it's a mark when dialing
	ErrReused = errors.New("Conn is reused")
)
