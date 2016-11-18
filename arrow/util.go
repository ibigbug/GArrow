package arrow

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

func debug(f string, a ...interface{}) {
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

func pipeConnWithContext(ctx context.Context, dst, src net.Conn) {
	var canceled = false
	var c = make(chan int, 1)
	go func() {
		for !canceled {
			buf := make([]byte, 1024)
			n, err := src.Read(buf)
			if n > 0 {
				dst.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		c <- 1
	}()
	select {
	case <-ctx.Done():
		canceled = true
	case <-c:
		debug("pipe done")
	}
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
