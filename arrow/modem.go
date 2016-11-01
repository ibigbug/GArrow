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

func Dial(network, remote string) (c net.Conn, err error) {
	return net.Dial(network, remote)
}

var ArrowTransport = &http.Transport{
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return Dial(network, address)
	},
	DisableKeepAlives:     false,
	DisableCompression:    false,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   100,
	ResponseHeaderTimeout: 5 * time.Second,
}
