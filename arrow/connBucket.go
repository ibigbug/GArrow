package arrow

import "net"

var connBucket = make(chan net.Conn, 10)

func getFreeConn() (c net.Conn) {
	select {
	case c = <-connBucket:
	default:
	}
	return
}

func putFreeConn(c net.Conn) bool {
	select {
	case connBucket <- c:
		return true
	default:
		return false
	}
}
