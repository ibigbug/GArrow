package arrow

import (
	"crypto/aes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	IDLE_TIMEOUT = 60 * time.Second
)

func NewEncryptConn(conn net.Conn, cipher *Cipher, timeout time.Duration) (c *EncryptConn) {
	c = &EncryptConn{
		Conn:    conn,
		timeout: timeout,
		mu:      &sync.Mutex{},
		cipher:  cipher,
	}
	if timeout > 0 {
		c.SetDeadline(time.Now().Add(timeout))
	}
	return
}

type EncryptConn struct {
	net.Conn
	timeout time.Duration
	closed  bool
	reused  bool
	mu      *sync.Mutex
	cipher  *Cipher
}

func (c *EncryptConn) Read(b []byte) (n int, err error) {
	if c.cipher.decer == nil {
		iv := make([]byte, aes.BlockSize)
		n, err := io.ReadFull(c.Conn, iv)
		if n != aes.BlockSize || err != nil {
			err = fmt.Errorf("Error read cipher: %d, %s", n, err)
			return 0, err
		}
		c.cipher.initDecer(iv)
	}
	n, err = c.Conn.Read(b)
	if err != nil {
		if nerr, ok := err.(net.Error); ok {
			if !nerr.Temporary() || nerr.Timeout() {
				c.Close()
				return
			}
		}
	}
	if c.timeout > 0 {
		c.SetDeadline(time.Now().Add(c.timeout))
	}
	c.cipher.Decrypt(b[:n])
	return
}

func (c *EncryptConn) Write(b []byte) (n int, err error) {
	if c.cipher.encer == nil {
		iv := c.cipher.initEncer()
		c.Conn.Write(iv)
	}

	n, err = c.Conn.Write(c.cipher.Encrypt(b))
	if err != nil {
		if nerr, ok := err.(net.Error); ok {
			if !nerr.Temporary() || nerr.Timeout() {
				c.Close()
				return
			}
		}
	}
	if c.timeout > 0 {
		c.SetDeadline(time.Now().Add(c.timeout))
	}
	return
}

func (c *EncryptConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return c.Conn.Close()
}

func (c *EncryptConn) SetTimeout(t time.Duration) {
	c.timeout = t
	if t > 0 {
		c.SetDeadline(time.Now().Add(t))
	}
}

func (c *EncryptConn) String() string {
	return fmt.Sprintf("conn: %s <-> %s", c.LocalAddr(), c.RemoteAddr())
}
