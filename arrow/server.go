package arrow

import (
	"encoding/binary"
	"time"

	"net"

	"github.com/Sirupsen/logrus"
	raven "github.com/getsentry/raven-go"
	"github.com/ibigbug/conn-pool/connpool"
)

// Server struct
type Server struct {
	*Config
	logger   *logrus.Logger
	connPool *connpool.ConnectionPool
}

// Run new server
func (s *Server) Run() (err error) {
	if s.ServerAddress == "" {
		s.logger.Fatal("config.server can not be nil")
	}

	l, err := ArrowListen("tcp4", s.ServerAddress, s.Password)
	defer l.Close()
	checkError(err)

	s.logger.Infoln("Server running at: ", s.ServerAddress)
	for {
		conn, err := l.Accept()
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			s.logger.Errorln("Accept error: ", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(cConn net.Conn) {
	defer cConn.Close()
	var rConn *connpool.ManagedConn
	rHost, err := s.peekHeader(cConn)
	s.logger.Infoln("rHost got:", rHost)
	if err != nil {
		cConn.Close() // no leak
		s.logger.Errorln("Error reading header: ", err)
		return
	}
	rConn, err = s.connPool.GetTimeout(rHost, 5*time.Second)
	if err != nil {
		// 'cause io.Copy not started yet
		// Read/Write Deadline doesn't cover this case
		cConn.Close() // no leak
		raven.CaptureErrorAndWait(err, nil)
		s.logger.Errorln("Error dialing to remote: ", err)
		return
	}
	go pipeConn(cConn, rConn)
	pipeConn(rConn, cConn)
	// TODO: may reuse conn here
	s.connPool.Remove(rConn)
}

func (s *Server) peekHeader(conn net.Conn) (host string, err error) {
	var size int64
	err = binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			raven.CaptureErrorAndWait(err.(error), nil)
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(size))
			s.logger.Errorln("Got wrong header:", b)
		}
	}()

	header := make([]byte, size)
	conn.Read(header)
	host = string(header[:])
	return
}

// NewServer proxy server factory
func NewServer(c *Config) (s Runnable) {
	var logger = getLogger("server")

	connPool := connpool.NewPool()
	s = &Server{
		Config:   c,
		logger:   logger,
		connPool: &connPool,
	}
	return
}
