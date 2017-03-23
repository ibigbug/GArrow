package arrow

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"io"

	"github.com/Sirupsen/logrus"
	raven "github.com/getsentry/raven-go"
)

func init() {
	raven.SetDSN("https://a680a8b9cc344eb79f065cdeaaf490bd:dfe4e7fbcaba43d994743707cd27d0e4@sentry.io/151139")
}

func debug(f string, a ...interface{}) {
	if os.Getenv("ARROW_DEBUG") == "" {
		return
	}
	if len(a) == 0 && !strings.HasSuffix(f, "\n") {
		fmt.Println(f)
	} else {
		fmt.Printf(f, a...)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func setHost(rConn net.Conn, rHost string) (err error) {
	rHost = ensurePort(rHost)
	err = binary.Write(rConn, binary.LittleEndian, int64(len(rHost)))
	if err != nil {
		return
	}
	// TODO: byte order?
	_, err = rConn.Write([]byte(rHost))
	return
}

func pipeConn(dst, src net.Conn) {
	r, w := io.Pipe()
	go func() {
		io.Copy(dst, r)
	}()
	io.Copy(w, src)
}

func ensurePort(s string) (h string) {
	h = s
	if !strings.Contains(s, ":") {
		h = fmt.Sprintf("%s:80", s)
	}
	return
}

func writeHeader(w http.ResponseWriter, hdr http.Header) {
	for k, vv := range hdr {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}

func getLogger(name string) *logrus.Logger {
	var logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	logger.Level = logrus.DebugLevel
	logger.WithFields(logrus.Fields{
		"from": "client",
	})
	return logger
}
