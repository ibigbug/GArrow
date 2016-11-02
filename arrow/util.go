package arrow

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	ErrIdle = errors.New("Socket idle too long")
)

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

func pipeConn(rConn io.ReadWriter, cConn io.ReadWriter) {

	go func() {
		r1, w1 := io.Pipe()
		go io.Copy(rConn, r1)
		io.Copy(w1, cConn)
	}()

	r2, w2 := io.Pipe()
	go io.Copy(cConn, r2)
	io.Copy(w2, rConn)
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
