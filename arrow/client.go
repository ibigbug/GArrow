package arrow

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
)

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

type ProxyHandler struct {
	serverAddr string
	logger     *logrus.Logger
	password   string
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Infoln(r.Method, r.URL.Path, r.Proto)

	if r.Method == "CONNECT" {
		rConn, err := Dial("tcp4", h.serverAddr, h.password)
		if err != nil {
			fmt.Fprintln(w, "Error connecting proxy server: ", err)
			return
		}

		err = setHost(rConn, r.Host)
		if err != nil {
			fmt.Fprintln(w, "Error negotiating with proxy server", err)
			return
		}

		hj, ok := w.(http.Hijacker)
		if !ok {
			fmt.Fprintln(w, "Error doing proxy hijack:", http.StatusInternalServerError)
			return
		}
		// TODO(ibigbug): ignore the hijack buf
		// If meet strange packet lossing issue, check here
		cConn, _, err := hj.Hijack()
		defer cConn.Close()
		if err != nil {
			fmt.Fprintln(w, "Error doing proxy hijack:", err)
			return
		}
		cConn.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
		pipeConn(rConn, cConn)
	} else {
		defer r.Body.Close()
		h.preprocessHeader(r)
		var d = &map[string]string{
			"password": h.password,
			"rHost":    r.Host,
			"address":  h.serverAddr,
		}
		var ctx = context.WithValue(r.Context(), "d", d)
		res, err := ArrowTransport.RoundTrip(r.WithContext(ctx))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "Error proxy request:", err)
			return
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		writeHeader(w, res.Header)
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
		// TODO(ibigbug) Trailers
	}
}

func (h *ProxyHandler) preprocessHeader(r *http.Request) {
	for _, h := range hopHeaders {
		r.Header.Del(h)
	}
}

type Client struct {
	*Config
	logger *logrus.Logger
}

func (c *Client) Run() (err error) {
	h := &ProxyHandler{
		serverAddr: c.ServerAddress,
		logger:     c.logger,
		password:   c.Password,
	}

	s := http.Server{
		Handler: h,
	}

	var l net.Listener
	if l, err = net.Listen("tcp", c.LocalAddress); err != nil {
		return err
	}
	c.logger.Infoln("Running client at: ", c.LocalAddress)
	return s.Serve(l)
}

func NewClient(c *Config) (s Runnable) {
	var logger = getLogger("client")
	s = &Client{
		Config: c,
		logger: logger,
	}
	return
}
