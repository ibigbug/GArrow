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

func NewArrowConn(conn net.Conn, cipher *Cipher, timeout time.Duration) (c *ArrowConn) {
	c = &ArrowConn{
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

type ArrowConn struct {
	net.Conn
	timeout time.Duration
	closed  bool
	reused  bool
	mu      *sync.Mutex
	cipher  *Cipher
	disable bool
}

func (c *ArrowConn) Read(b []byte) (n int, err error) {
	if !c.disable {
		if c.cipher.decer == nil {
			iv := make([]byte, aes.BlockSize)
			n, err := io.ReadFull(c.Conn, iv)
			if n != aes.BlockSize || err != nil {
				err = fmt.Errorf("Error read cipher: %d, %s", n, err)
				return 0, err
			}
			c.cipher.initDecer(iv)
		}
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
	if !c.disable {
		c.cipher.Decrypt(b[:n])
	}
	return
}

func (c *ArrowConn) Write(b []byte) (n int, err error) {
	if !c.disable {
		if c.cipher.encer == nil {
			iv := c.cipher.initEncer()
			c.Conn.Write(iv)
		}
	}

	if !c.disable {
		n, err = c.Conn.Write(c.cipher.Encrypt(b))
	} else {
		n, err = c.Conn.Write(b)
	}
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

func (c *ArrowConn) SetTimeout(t time.Duration) {
	c.timeout = t
	if t > 0 {
		c.SetDeadline(time.Now().Add(t))
	}
}

func (c *ArrowConn) String() string {
	return fmt.Sprintf("conn: %s <-> %s", c.LocalAddr(), c.RemoteAddr())
}

func (c *ArrowConn) DisableEncrypt(b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.disable = b
}
