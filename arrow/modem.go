package arrow

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Conn struct {
	net.Conn
}

func Dial(network, remote, password string) (c net.Conn, err error) {
	return net.Dial(network, remote)
}

var ArrowTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		password := ctx.Value("password").(string)
		return Dial(network, address, password)
	},
	DisableKeepAlives:     false,
	DisableCompression:    false,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   100,
	ResponseHeaderTimeout: 5 * time.Second,
}
