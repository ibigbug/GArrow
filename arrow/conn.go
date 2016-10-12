package arrow

import (
	"net"
	"sync"
)

type HeaderedConn struct {
	net.Conn
	heanderSent bool
	mu          *sync.Mutex
}

type ArrowListener struct {
	rl net.Listener // the raw listener
}

func ArrowListen(_net, laddr string) (l net.Listener, err error) {
	rl, err := net.Listen(_net, laddr)
	if err != nil {
		return
	}
	l = &ArrowListener{
		rl: rl,
	}
	return
}

func (l *ArrowListener) Accept() (c net.Conn, err error) {
	rc, err := l.rl.Accept()
	if err != nil {
		return
	}
	c = &HeaderedConn{
		Conn:        rc,
		heanderSent: false,
		mu:          &sync.Mutex{},
	}
	return
}

func (l *ArrowListener) Close() error {
	return l.rl.Close()
}

func (l *ArrowListener) Addr() net.Addr {
	return l.rl.Addr()
}
