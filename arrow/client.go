package arrow

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/ibigbug/conn-pool/connpool"
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
	connPool   *connpool.ConnectionPool
	logger     *logrus.Logger
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugln("new req:", r.Method, r.URL, r.Proto)

	if r.Method == "CONNECT" {
		h.preprocessHeader(r)
	}

	rConn, err := h.connPool.Get(h.serverAddr)
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
		return
	}

	h.writeConnHeader(rConn, r, w)
	hj, ok := w.(http.Hijacker)
	if !ok {
		fmt.Fprintln(w, "Error doing proxy hijack:", http.StatusInternalServerError)
		return
	}
	cConn, buf, err := hj.Hijack()
	defer cConn.Close()
	if err != nil {
		fmt.Fprintln(w, "Error doing proxy hijack:", err)
		return
	}

	// WTF the difference between httputil.DumpRequest ?
	req, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		fmt.Fprintln(w, "Error doing proxy hijack:", err)
	}
	rConn.Write(req)
	bs := buf.Reader.Buffered()
	if bs > 0 {
		remain := make([]byte, bs)
		n, _ := buf.Read(remain)
		if n != bs {
			h.logger.Warningln(n, bs)
			fmt.Fprintln(w, "Request malformed", http.StatusBadRequest)
		}
		rConn.Write(remain)
	}

	if r.Method == "CONNECT" {
		cConn.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))

	}
	pipeWithTimeout(rConn, cConn)
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
	cp := connpool.NewPool()
	h := &ProxyHandler{
		serverAddr: c.ServerAddress,
		connPool:   &cp,
		logger:     c.logger,
	}

	s := http.Server{
		Handler: h,
		ConnState: func(c net.Conn, s http.ConnState) {
			fmt.Println("conn:", c.LocalAddr(), "<->", c.RemoteAddr(), "state:", s)
		},
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
