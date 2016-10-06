package arrow

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func makeConnChan(conn io.ReadWriter) (c chan []byte) {
	c = make(chan []byte)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("Error recover: %v", r)
				}
				log.Println("Recovered from chan:", err)
			}
		}()
		b := make([]byte, 1024)
		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				c <- res
			}

			if err != nil {
				c <- nil
				if err != io.EOF {
					log.Println("piping error:", err)
				}
			}
		}
	}()

	return
}

func pipeWithTimeout(rConn *net.TCPConn, cConn io.ReadWriter) {

	rChan := makeConnChan(rConn)
	cChan := makeConnChan(cConn)

	timeout := make(chan bool, 1)
	t := time.NewTimer(10 * time.Second)
	go func() {
		<-t.C
		timeout <- true
	}()

	for {
		select {
		case b1 := <-rChan:
			if !t.Stop() {
				<-t.C
			}
			if b1 == nil {
				return
			}
			cConn.Write(b1)
			t.Reset(10 * time.Second)
		case b2 := <-cChan:
			if !t.Stop() {
				<-t.C
			}
			if b2 == nil {
				return
			}
			rConn.Write(b2)
			t.Reset(10 * time.Second)

		case <-timeout:
			close(rChan)
			close(cChan)
			break
		}
	}
}

func ensurePort(s string) (h string) {
	h = s
	if !strings.Contains(s, ":") {
		h = fmt.Sprintf("%s:80", s)
	}
	return
}
