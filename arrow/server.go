package arrow

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
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

	tcpAddr, err := net.ResolveTCPAddr("tcp", s.ServerAddress)
	if err != nil {
		s.logger.Fatalln("Invalid config.server: ", s.ServerAddress, err)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	defer l.Close()
	checkError(err)

	s.logger.Infoln("Server running at: ", s.ServerAddress)
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			s.logger.Errorln("Accept error: ", err)
			continue
		}
		go s.handle(&ConnWithHeader{
			TCPConn:      conn,
			headerPeeked: false,
		})
	}
}

func (s *Server) handle(cConn *ConnWithHeader) {
	var rConn *connpool.ManagedConn

	rHost, err := s.peekHeader(cConn)
	fmt.Println("rHost got:", rHost)
	if err != nil {
		s.logger.Errorln("Error reading header: ", err)
		return
	}

	rConn, err = s.connPool.GetTimeout(rHost, 5*time.Second)
	if err != nil {
		s.logger.Errorln("Error dialing to remote: ", err)
		return
	}

	pipeWithTimeout(rConn, cConn)
	s.connPool.Remove(rConn)

}

func (s *Server) peekHeader(conn *ConnWithHeader) (host string, err error) {
	var size int64
	err = binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		return
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(size))
	header := make([]byte, size)
	conn.Read(header)
	host = string(header[:])
	return
}

// NewServer proxy server factory
func NewServer(c *Config) (s Runnable) {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stderr)
	var logger = logrus.New()
	logger.WithFields(logrus.Fields{
		"from": "server",
	})
	connPool := connpool.NewPool()
	s = &Server{
		Config:   c,
		logger:   logger,
		connPool: &connPool,
	}
	return
}
