package arrow

import "net"

type Runnable interface {
	Run() error
}

type ConnWithHeader struct {
	*net.TCPConn
	headerPeeked bool
}
