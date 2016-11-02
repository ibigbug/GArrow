package arrow

import (
	"context"
	"crypto/aes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	IDLE_TIMEOUT = 5 * time.Second
)

type Conn struct {
	net.Conn
	timeout time.Duration
	cipher  *Cipher
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.cipher.decer == nil {
		iv := make([]byte, aes.BlockSize)
		n, err := io.ReadFull(c.Conn, iv)
		if n != aes.BlockSize || err != nil {
			err = fmt.Errorf("Error read cipher: %d, %s", n, err)
			return -1, err
		}
		c.Conn.Write(iv)
		c.cipher.initDecer(iv)
	}
	n, err = c.Conn.Read(b)
	c.cipher.Decrypt(b)
	if c.timeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.cipher.encer == nil {
		c.cipher.initEncer()
	}
	n, err = c.Conn.Write(c.cipher.Encrypt(b))
	if c.timeout > 0 {
		c.Conn.SetDeadline(time.Now().Add(c.timeout))
	}
	return
}

func (c *Conn) SetDeadLine(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func Dial(network, remote, password string) (c net.Conn, err error) {
	var rc net.Conn
	rc, err = net.Dial(network, remote)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting proxy server", err)
		return
	}
	cipher, err := NewCipher(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting cipher", err)
		return
	}
	c = &Conn{
		Conn:    rc,
		timeout: IDLE_TIMEOUT, // idle timeout
		cipher:  cipher,
	}
	c.SetDeadline(time.Now().Add(IDLE_TIMEOUT))
	return
}

var ArrowTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		password := ctx.Value("password").(string)
		return Dial(network, address, password)
	},
	DisableKeepAlives:     false,
	DisableCompression:    false,
	MaxIdleConns:          10,
	MaxIdleConnsPerHost:   10,
	ResponseHeaderTimeout: 5 * time.Second,
}

type ArrowListener struct {
	net.Listener
	cipher *Cipher
}

func (l *ArrowListener) Accept() (c net.Conn, err error) {
	rc, err := l.Listener.Accept()
	c = &Conn{
		Conn:    rc,
		cipher:  l.cipher,
		timeout: 0,
	}
	return
}

func ArrowListen(network, address, password string) (l net.Listener, err error) {
	cipher, err := NewCipher(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting cipher", err)
		return
	}

	rl, err := net.Listen(network, address)
	l = &ArrowListener{
		Listener: rl,
		cipher:   cipher,
	}
	return
}
