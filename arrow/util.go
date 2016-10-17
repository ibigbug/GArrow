package arrow

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrIdle = errors.New("Socket idle too long")
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func pipeWithTimeout(rConn io.ReadWriter, cConn io.ReadWriter) {

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
