package ioUtils

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/def"
	"github.com/tony-507/analyzers/src/plugins/ioUtils/protocol"
	"golang.org/x/net/ipv4"
)

// Assume UDP protocol
type sockConn struct {
	logger  common.Log
	address string
	port    string
	itf     string
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
	n, _, _, err := s.conn.ReadFrom(buf)
	if err != nil {
		s.logger.Fatal("Cannot read from UDP datagram")
		return buf, err
	}
	return buf[:n], nil
}

func socketConnection(logger common.Log, address string, port string, itf string) *sockConn {
	return &sockConn{logger: logger, address: address, port: port, itf: itf, conn: nil}
}

type udpReaderStruct struct {
	logger      common.Log
	address     string
	port        string
	itf         string
	conn        *sockConn
	bufferQueue []def.ParseResult
	udpCount    int
	config      def.IReaderConfig
}

func (ur *udpReaderStruct) Setup(config def.IReaderConfig) {
	ur.conn = socketConnection(ur.logger, ur.address, ur.port, ur.itf)
	ur.config = config
}

func (ur *udpReaderStruct) StartRecv() error {
	return ur.conn.init()
}

func (ur *udpReaderStruct) StopRecv() error {
	return ur.conn.close()
}

func (ur *udpReaderStruct) DataAvailable() (def.ParseResult, bool) {
	if len(ur.bufferQueue) <= 1 {
		udpBuf, err := ur.conn.read()

		if err != nil {
			ur.logger.Error("Fail to read buffer: %s", err.Error())
		}

		ur.udpCount += 1

		ur.bufferQueue = append(ur.bufferQueue, protocol.ParseWithParsers(ur.config.Protocols, udpBuf)...)
	}

	buf := ur.bufferQueue[0]
	ur.bufferQueue = ur.bufferQueue[1:]

	return buf, true
}

func udpReader(param *udpInputParam, name string) def.IReader {
	rv := udpReaderStruct{}
	rv.conn = nil
	rv.udpCount = 0

	rv.logger = common.CreateLogger(name)
	rv.bufferQueue = make([]def.ParseResult, 0)

	tmp := strings.Split(param.Address, ":")
	rv.address = tmp[0]
	rv.port = tmp[1]
	rv.itf = param.Itf

	return &rv
}
