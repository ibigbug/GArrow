package arrow

import (
	"encoding/binary"
	"log"
	"net"
	"os"

	"github.com/Sirupsen/logrus"
)

type Server struct {
	*Config
	logger *logrus.Logger
}

func (s *Server) Run() (err error) {
	if s.ServerAddress == "" {
		log.Fatal("config.server can not be nil")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", s.ServerAddress)
	if err != nil {
		log.Fatalln("Invalid config.server: ", s.ServerAddress, err)
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
		go s.HandleConn(&ConnWithHeader{
			TCPConn:      conn,
			headerPeeked: false,
		})
	}
}

func (s *Server) HandleConn(cConn *ConnWithHeader) {
	var rConn *net.TCPConn
	defer cConn.Close()

	for {
		if !cConn.headerPeeked {
			// block wating header for next client request reuese this conn
			rHost, err := s.peekHeader(cConn)
			if err != nil {
				s.logger.Errorln("Error reading header: ", err)
				return
			}
			tcpAddr, err := net.ResolveTCPAddr("tcp", rHost)
			if err != nil {
				s.logger.Errorln("Error resolve remote addr: ", err)
				return
			}

			// we close it manually
			rConn, err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				s.logger.Errorln("Error dialing to remote: ", err)
				return
			}
			pipeWithTimeout(rConn, cConn)
			rConn.Close()
			cConn.headerPeeked = false
		}
	}
}

func (s *Server) peekHeader(conn *ConnWithHeader) (host string, err error) {
	var size int64
	err = binary.Read(conn, binary.LittleEndian, &size)
	if err != nil {
		return
	}
	header := make([]byte, size)
	conn.Read(header)
	host = string(header[:])
	return
}

func NewServer(c *Config) (s Runnable) {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stderr)
	logger := logrus.New()

	s = &Server{
		Config: c,
		logger: logger,
	}
	return
}
