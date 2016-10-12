package arrow

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
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

type ProxyHandler struct {
	tcpAddr *net.TCPAddr
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new req:", r.Method, r.URL, r.Proto)
	ph.preprocessHeader(r)
	if r.Method == "CONNECT" {
		ph.handleHTTPS(w, r)
	} else {
		ph.handleHTTP(w, r)
	}
}

func (ph *ProxyHandler) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	rConn, err := net.DialTCP("tcp", nil, ph.tcpAddr)
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
		return
	}

	// tell remote connection header info
	// remoteAddr,
	ph.writeConnHeader(rConn, r, w)

	hj, ok := w.(http.Hijacker)
	if !ok {
		fmt.Fprintln(w, "Error doing proxy hijack:", http.StatusInternalServerError)
		return
	}
	cConn, buf, err := hj.Hijack()

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
			log.Fatalln(n, bs)
		}
		rConn.Write(remain)
	}
	cConn.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
	pipeWithTimeout(rConn, cConn)
}

func (ph *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	rConn, err := net.DialTCP("tcp", nil, ph.tcpAddr)
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
		return
	}

	ph.writeConnHeader(rConn, r, w)
	fmt.Println("header sent")
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
			log.Fatalln(n, bs)
		}
		rConn.Write(remain)
	}
	fmt.Println("req header sent")
	pipeWithTimeout(rConn, cConn)
}

func (ph *ProxyHandler) preprocessHeader(r *http.Request) {
	for _, h := range hopHeaders {
		r.Header.Del(h)
	}
}

func (ph *ProxyHandler) writeConnHeader(rConn io.ReadWriter, r *http.Request, w http.ResponseWriter) {
	rHost := ensurePort(r.Host)
	err := binary.Write(rConn, binary.LittleEndian, int64(len(rHost)))
	fmt.Println("rHost:", rHost, "size:", len(rHost))
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}

	// TODO: byte order?
	n, err := rConn.Write([]byte(rHost))
	if n != len(rHost) {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}
}

type Client struct {
	*Config
	logger *logrus.Logger
}

func (c *Client) Run() error {

	tcpAddr, err := net.ResolveTCPAddr("tcp", c.ServerAddress)
	if err != nil {
		c.logger.Errorln("Error resolving proxy server address: ", err)
		os.Exit(1)
	}
	h := &ProxyHandler{
		tcpAddr: tcpAddr,
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
