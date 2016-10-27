package connpool

import "sync"

type ConnManager struct {
	sync.Mutex
	conns []*ManagedConn
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make([]*ManagedConn, 0),
	}
}
