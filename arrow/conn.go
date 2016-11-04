package arrow

import (
	"crypto/aes"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	IDLE_TIMEOUT = 5 * time.Second
)

func NewTimeoutConn(conn net.Conn, timeout time.Duration) (c *TimeoutConn) {
	c = &TimeoutConn{
		Conn:    conn,
		timeout: timeout,
	}
	if timeout > 0 {
		c.SetDeadline(time.Now().Add(timeout))
	}
	return
}

type TimeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (n int, err error) {
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
		c.Conn.SetDeadline(time.Now().Add(c.timeout))
	}
	return
}

func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if err != nil {
		if nerr, ok := err.(net.Error); ok {
			if !nerr.Temporary() || nerr.Timeout() {
				c.Close()
				return
			}
		}
	}
	if c.timeout > 0 {
		c.Conn.SetDeadline(time.Now().Add(c.timeout))
	}
	return
}

func (c *TimeoutConn) SetTimeout(t time.Duration) {
	c.timeout = t
	if t > 0 {
		c.Conn.SetDeadline(time.Now().Add(t))
	}
}

func NewEncryptConn(conn net.Conn, cipher *Cipher) (c net.Conn) {
	tc := NewTimeoutConn(conn, IDLE_TIMEOUT)
	c = &EncryptConn{
		TimeoutConn: tc,
		cipher:      cipher,
	}
	return
}

type EncryptConn struct {
	*TimeoutConn
	cipher *Cipher
}

func (c *EncryptConn) Read(b []byte) (n int, err error) {
	if c.cipher.decer == nil {
		iv := make([]byte, aes.BlockSize)
		n, err := io.ReadFull(c.TimeoutConn, iv)
		if n != aes.BlockSize || err != nil {
			err = fmt.Errorf("Error read cipher: %d, %s", n, err)
			return 0, err
		}
		c.cipher.initDecer(iv)
	}
	n, err = c.TimeoutConn.Read(b)
	c.cipher.Decrypt(b[:n])
	return
}

func (c *EncryptConn) Write(b []byte) (n int, err error) {
	if c.cipher.encer == nil {
		iv := c.cipher.initEncer()
		c.TimeoutConn.Write(iv)
	}

	n, err = c.TimeoutConn.Write(c.cipher.Encrypt(b))
	return
}
