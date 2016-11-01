package arrow

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

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

// ProxyHandler handle requests
type ProxyHandler struct {
	serverAddr string
	logger     *logrus.Logger
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Infoln("new req:", r.Method, r.URL.Path, r.Proto)

	rConn, err := Dial("tcp4", h.serverAddr)
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
		return
	}

	h.writeConnHeader(rConn, r, w)

	if r.Method == "CONNECT" {
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
		rConn.Close() // TODO: using pool
	} else {
		h.preprocessHeader(r)
		res, err := ArrowTransport.RoundTrip(r)
		defer res.Body.Close()
		if err != nil {
			fmt.Fprintln(w, "Error proxy request:", err)
			return
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

func (h *ProxyHandler) writeConnHeader(rConn io.ReadWriter, r *http.Request, w http.ResponseWriter) {
	rHost := ensurePort(r.Host)
	err := binary.Write(rConn, binary.LittleEndian, int64(len(rHost)))
	h.logger.Debugln("rHost:", rHost, "size:", len(rHost))
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}

	// TODO: byte order?
	n, err := rConn.Write([]byte(rHost))
	if n != len(rHost) {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}
}

// Client definition
type Client struct {
	*Config
	logger *logrus.Logger
}

// Run proxy client
func (c *Client) Run() (err error) {
	h := &ProxyHandler{
		serverAddr: c.ServerAddress,
		logger:     c.logger,
	}

	s := http.Server{
		Handler: h,
	}

	l, err := net.Listen("tcp", c.LocalAddress)
	c.logger.Infoln("Running client at: ", c.LocalAddress)

	return s.Serve(l)
}

// NewClient factory
func NewClient(c *Config) (s Runnable) {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stderr)
	var logger = logrus.New()
	logger.WithFields(logrus.Fields{
		"from": "client",
	})

	s = &Client{
		Config: c,
		logger: logger,
	}
	return
}
