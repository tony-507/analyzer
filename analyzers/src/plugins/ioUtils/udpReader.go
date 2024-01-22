package ioUtils

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/plugins/common/protocol"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
	"golang.org/x/net/ipv4"
)

// Assume UDP protocol
type sockConn struct {
	logger  logging.Log
	address string
	port    string
	itf     string
	timeout int
	conn    *ipv4.PacketConn
}

func (s *sockConn) close() error {
	return s.conn.Close()
}

func (s *sockConn) init() error {
	port, err := strconv.Atoi(s.port)
	if err != nil {
		s.logger.Fatal("Port is not an integer: %s", s.port)
		return errors.New("Invalid port")
	}

	a := net.ParseIP(s.address)
	if a == nil {
		s.logger.Fatal("Fail to parse IP address %s", s.address)
		return errors.New("Bad IP address")
	}

	return s.openMcast(a, port)
}

func (s *sockConn) openMcast(a net.IP, port int) error {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return err
	}

	err = syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}

	err = syscall.SetsockoptString(sock, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, s.itf)
	if err != nil {
		return err
	}

	lsa := syscall.SockaddrInet4{Port: port}
	copy(lsa.Addr[:], a.To4())

	err = syscall.Bind(sock, &lsa)
	if err != nil {
		syscall.Close(sock)
		return err
	}
	f := os.NewFile(uintptr(sock), "")
	c, err := net.FilePacketConn(f)
	f.Close()
	if err != nil {
		return err
	}
	p := ipv4.NewPacketConn(c)

	itfName, err := net.InterfaceByName(s.itf)
	if err != nil {
		return err
	}

	err = p.JoinGroup(itfName, &net.UDPAddr{IP: a})
	if err != nil {
		return err
	}

	err = p.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	if err != nil {
		return err
	}

	s.conn = p
	return nil
}

func (s *sockConn) read() ([]byte, error) {
	buf := make([]byte, 10000)
	setTimeoutErr := s.conn.SetDeadline(time.Now().Add(time.Duration(s.timeout) * time.Second))
	if setTimeoutErr != nil {
		return buf, setTimeoutErr
	}
	n, _, _, err := s.conn.ReadFrom(buf)
	if err != nil {
		return buf, err
	}
	return buf[:n], nil
}

func socketConnection(logger logging.Log, address string, port string, itf string, timeout int) *sockConn {
	return &sockConn{
		logger: logger,
		address: address,
		port: port,
		itf: itf,
		conn: nil,
		timeout: timeout,
	}
}

type udpReaderStruct struct {
	logger      logging.Log
	address     string
	port        string
	itf         string
	timeout     int
	conn        *sockConn
	bufferQueue []protocol.ParseResult
	udpCount    int
	config      def.IReaderConfig
}

func (ur *udpReaderStruct) Setup(config def.IReaderConfig) {
	ur.conn = socketConnection(ur.logger, ur.address, ur.port, ur.itf, ur.timeout)
	ur.config = config
}

func (ur *udpReaderStruct) StartRecv() error {
	return ur.conn.init()
}

func (ur *udpReaderStruct) StopRecv() error {
	return ur.conn.close()
}

func (ur *udpReaderStruct) DataAvailable() (protocol.ParseResult, bool) {
	if len(ur.bufferQueue) <= 1 {
		udpBuf, err := ur.conn.read()

		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "i/o timeout") {
				panic(err)
			}
			ur.logger.Error("Fail to read buffer: %s", msg)
		}

		ur.udpCount += 1

		ur.bufferQueue = append(ur.bufferQueue, protocol.ParseWithParsers(ur.config.Parsers, &protocol.ParseResult{Buffer: udpBuf})...)
	}

	buf := ur.bufferQueue[0]
	ur.bufferQueue = ur.bufferQueue[1:]

	return buf, true
}

func udpReader(param *udpInputParam, name string) def.IReader {
	rv := udpReaderStruct{}
	rv.conn = nil
	rv.udpCount = 0

	rv.logger = logging.CreateLogger(name)
	rv.bufferQueue = make([]protocol.ParseResult, 0)

	tmp := strings.Split(param.Address, ":")
	rv.address = tmp[0]
	rv.port = tmp[1]
	rv.itf = param.Itf
	if param.Timeout > 0 {
		rv.timeout = param.Timeout
	} else {
		rv.timeout = 3
	}

	return &rv
}
