package connpool

import (
	"fmt"
	"net"
	"strings"
	"sync"
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
		pool:    make(map[string][]*ManagedConn),
		mutex:   &sync.Mutex{}, // should lock per host
	}
}

// ConnectionPool a connection pool
type ConnectionPool struct {
	timeout time.Duration
	pool    map[string][]*ManagedConn
	mutex   *sync.Mutex
}

// SetKeepAliveTimeout sets after `to` seconds, conn released to pool will be removed
func (p *ConnectionPool) SetKeepAliveTimeout(to time.Duration) {
	p.timeout = to
}

// Get get a connection with specific remote address
// could be domain:port/ip:port
func (p ConnectionPool) Get(remoteAddr string) (conn *ManagedConn, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	remoteAddr = ensurePort(remoteAddr)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return
	}
	id := tcpAddr.String()
	conns := p.pool[id]
	if conns == nil {
		p.pool[id] = make([]*ManagedConn, 0)
	}

	if len(conns) == 0 {
		conn = p.createConn(tcpAddr)
	} else {
		conn = p.getFreeConn(id)
		if conn == nil {
			conn = p.createConn(tcpAddr)
		} else {
			debug("reusing conn %p, idle changed to 0\n", conn)
			atomic.StoreUint32(&conn.idle, 0)
		}
	}
	return
}

// Release release to pool after used, and will be closed in `timeout` seconds
func (p ConnectionPool) Release(conn *ManagedConn) {
	debug("releasing conn %p, idle changed to 1\n", conn)
	atomic.StoreUint32(&conn.idle, 1)

	go func() {
		timer := time.NewTimer(p.timeout)
		<-timer.C
		remoteAddr := conn.RemoteAddr().String()

		// Lock it
		p.mutex.Lock()
		defer p.mutex.Unlock()

		if atomic.LoadUint32(&conn.idle) == 0 {
			debug("conn %p is reused, skipping release\n", conn)
			return
		}
		debug("doing release %p(idle: %v), current conn pool: ", conn, conn.idle)
		for _, c := range p.pool[remoteAddr] {
			debug("%p, ", c)
		}
		debug("")
		idx := findIdx(p.pool[remoteAddr], conn)

		if idx == -1 {
			// conn has already been released
			return
		}
		p.pool[remoteAddr] = append(p.pool[remoteAddr][:idx], p.pool[remoteAddr][idx+1:]...)
		conn.Close()
	}()
}

func (p ConnectionPool) createConn(tcpAddr *net.TCPAddr) (conn *ManagedConn) {
	rawConn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err == nil {
		conn = &ManagedConn{
			TCPConn: *rawConn,
			idle:    0,
		}
		p.pool[tcpAddr.String()] = append(p.pool[tcpAddr.String()], conn)
		debug("creating new conn: %p\n", conn)
	}
	return
}

func (p ConnectionPool) getFreeConn(tcpAddr string) (c *ManagedConn) {
	for _, c = range p.pool[tcpAddr] {
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
