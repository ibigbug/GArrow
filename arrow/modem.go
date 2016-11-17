package arrow

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func Dial(network, remote, password string, reuse bool) (c net.Conn, err error) {
	var rc net.Conn

	if reuse {
		c = getFreeConn()
		if c != nil && !c.(*EncryptConn).closed {
			debug("reuse conn")
			return
		} else {
			if c != nil {
				debug("reused conn closed, dial new one")
			}
		}
	}

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
	c = NewEncryptConn(rc, cipher, IDLE_TIMEOUT)
	return
}

var ArrowTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, _ string) (c net.Conn, err error) {
		d := ctx.Value("d").(*map[string]string)
		c, err = Dial(network, (*d)["address"], (*d)["password"], false)
		setHost(c, (*d)["rHost"])
		return
	},
	DisableKeepAlives:     false,
	DisableCompression:    false,
	MaxIdleConns:          10,
	MaxIdleConnsPerHost:   10,
	IdleConnTimeout:       5 * time.Second,
	ResponseHeaderTimeout: 5 * time.Second,
}

type ArrowListener struct {
	net.Listener
	password string
}

func (l *ArrowListener) Accept() (c net.Conn, err error) {
	cipher, err := NewCipher(l.password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting cipher", err)
		return
	}
	rc, err := l.Listener.Accept()
	c = NewEncryptConn(rc, cipher, IDLE_TIMEOUT)
	return
}

func ArrowListen(network, address, password string) (l net.Listener, err error) {
	rl, err := net.Listen(network, address)
	l = &ArrowListener{
		Listener: rl,
		password: password,
	}
	return
}
