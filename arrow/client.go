package arrow

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
)

type ProxyHandler struct {
	tcpAddr *net.TCPAddr
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ph.handleHTTP(w, r)
}

func (ph *ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {

	rConn, err := net.DialTCP("tcp", nil, ph.tcpAddr)
	defer rConn.Close()
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
		return
	}
	rHost := ensurePort(r.Host)
	err = binary.Write(rConn, binary.LittleEndian, int64(len(rHost)))
	if err != nil {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}

	// TODO: byte order?
	n, err := rConn.Write([]byte(rHost))
	if n != len(rHost) {
		fmt.Fprintln(w, "Error connecting proxy server: ", err)
	}

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

	bs := buf.Reader.Buffered()
	bs2 := buf.Writer.Buffered()
	fmt.Println("remaining:", bs, bs2)
	if bs > 0 {
		remain := make([]byte, bs)
		n, _ := buf.Read(remain)
		if n != bs {
			log.Fatalln(n, bs)
		}
		rConn.Write(remain)
	}
	pipeWithTimeout(rConn, cConn)
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
	ph := &ProxyHandler{
		tcpAddr: tcpAddr,
	}

	c.logger.Infoln("Running client at: ", c.LocalAddress)
	return http.ListenAndServe(c.LocalAddress, ph)
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
