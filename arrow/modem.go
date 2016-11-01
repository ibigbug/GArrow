package arrow

import (
	"context"
	"net"
	"net/http"
)

type Conn struct {
}

func Dial(remote string) (c net.Conn, err error) {
	return net.Dial("tcp", remote)
}

var ArrowTransport http.RoundTripper = &http.Transport{
	DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return net.Dial(network, address)
	},
}
