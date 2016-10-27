package connpool

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

// DefaultKeepAliveTimeout default keepalive timeout in seconds
const (
	DefaultKeepAliveTimeout = time.Second * 60
)

// ManagedConn net.TCPConn with idle mark
type ManagedConn struct {
	net.TCPConn
	idle uint32
}

// NewPool get a new ConnectionPool
func NewPool() ConnectionPool {
	return ConnectionPool{
		timeout: DefaultKeepAliveTimeout,
		// id -> ConnManager
		// id is like 1.2.3.4:90, the resolved remoteAddr
		pool: make(map[string]*ConnManager),
	}
}

// ConnectionPool a connection pool
type ConnectionPool struct {
	timeout time.Duration
	pool    map[string]*ConnManager
}

// SetKeepAliveTimeout sets after `to` seconds, conn released to pool will be removed
func (p *ConnectionPool) SetKeepAliveTimeout(to time.Duration) {
	p.timeout = to
}

// Get get a connection with specific remote address
func (p *ConnectionPool) Get(remoteAddr string) (*ManagedConn, error) {
	return p.get(remoteAddr, 0)
}

// Get get a connection with timeout
func (p *ConnectionPool) GetTimeout(remoteAddr string, timeout time.Duration) (*ManagedConn, error) {
	return p.get(remoteAddr, timeout)
}

func (p *ConnectionPool) get(remoteAddr string, timeout time.Duration) (conn *ManagedConn, err error) {
	remoteAddr = ensurePort(remoteAddr)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return
	}
	id := tcpAddr.String()
	mgr := p.pool[id]

	if mgr == nil {
		mgr = NewConnManager()
		p.pool[id] = mgr
	}

	mgr.Lock()
	defer mgr.Unlock()

	if len(mgr.conns) == 0 {
		conn, err = p.createConn(tcpAddr, timeout)
	} else {
		conn = p.getFreeConn(id)
		if conn == nil {
			conn, err = p.createConn(tcpAddr, timeout)
		} else {
			debug("reusing conn %p, idle changed to 0\n", conn)
			atomic.StoreUint32(&conn.idle, 0)
		}
	}
	return
}

// Remove the connection immediately
func (p *ConnectionPool) Remove(conn *ManagedConn) {
	debug("remove conn %p\n", conn)
	id := conn.RemoteAddr().String()
	mgr := p.pool[id]
	mgr.Lock()
	defer mgr.Unlock()
	p.remove(conn)
}

// Put release a conn back to pool, and will be removed in `timeout` seconds
func (p *ConnectionPool) Put(conn *ManagedConn) {
	debug("put back conn %p, idle changed to 1\n", conn)
	atomic.StoreUint32(&conn.idle, 1)

	go func() {
		timer := time.NewTimer(p.timeout)
		<-timer.C

		id := conn.RemoteAddr().String()
		mgr := p.pool[id]
		// Lock it
		mgr.Lock()
		defer mgr.Unlock()

		if atomic.LoadUint32(&conn.idle) == 0 {
			debug("conn %p is reused, skipping release\n", conn)
			return
		}
		p.remove(conn)
	}()
}

func (p *ConnectionPool) remove(conn *ManagedConn) {
	id := conn.RemoteAddr().String()
	mgr := p.pool[id]
	idx := findIdx(mgr.conns, conn)
	if idx == -1 {
		// conn has already been released
		return
	}
	mgr.conns = append(mgr.conns[:idx], mgr.conns[idx+1:]...)
	conn.Close()
}

func (p ConnectionPool) createConn(tcpAddr *net.TCPAddr, timeout time.Duration) (conn *ManagedConn, err error) {
	var rawConn net.Conn
	if timeout == 0 {
		rawConn, err = net.Dial("tcp4", tcpAddr.String())
	} else {
		rawConn, err = net.DialTimeout("tcp4", tcpAddr.String(), timeout)
	}
	if err == nil {
		conn = &ManagedConn{
			TCPConn: *rawConn.(*net.TCPConn),
			idle:    0,
		}
		mgr := p.pool[tcpAddr.String()]
		mgr.conns = append(mgr.conns, conn)
		debug("creating new conn: %p\n", conn)
	}
	return
}

func (p ConnectionPool) getFreeConn(id string) (c *ManagedConn) {
	for _, c = range p.pool[id].conns {
		debug("scanning cann %p, idle: %v\n", c, c.idle)
		if atomic.LoadUint32(&c.idle) == 1 {
			debug("found free conn: %p\n", c)
			return
		}
	}
	c = nil // dont return the last one
	return
}

func findIdx(arr []*ManagedConn, ele *ManagedConn) int {
	for idx, v := range arr {
		if v == ele {
			return idx
		}
	}
	return -1
}

func ensurePort(addr string) (rv string) {
	rv = addr
	if !strings.Contains(addr, ":") {
		rv = fmt.Sprintf("%s:80", rv)
	}
	return
}
