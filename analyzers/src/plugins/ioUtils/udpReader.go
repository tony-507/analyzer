package ioUtils

import (
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/tony-507/analyzers/src/common"
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

func (s *sockConn) close() {
	err := s.conn.Close()
	if err != nil {
		s.logger.Error("Cannot close socket connection: %s", err.Error())
	}
}

func (s *sockConn) init() {
	port, err := strconv.Atoi(s.port)
	if err != nil {
		s.logger.Fatal("Port is not an integer: %s", s.port)
		panic("Invalid port")
	}

	a := net.ParseIP(s.address)
	if a == nil {
		s.logger.Fatal("Fail to parse IP address %s", s.address)
		panic("Bad IP address")
	}

	s.openMcast(a, port)
}

func (s *sockConn) openMcast(a net.IP, port int) {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		s.logger.Fatal("Cannot open socket: %s", err.Error())
		panic(err)
	}

	err = syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		s.logger.Fatal("Cannot set socket option: %s", err.Error())
		panic(err)
	}

	err = syscall.SetsockoptString(sock, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, s.itf)
	if err != nil {
		s.logger.Fatal("Cannot set socket option: %s", err.Error())
		panic(err)
	}

	lsa := syscall.SockaddrInet4{Port: port}
	copy(lsa.Addr[:], a.To4())

	err = syscall.Bind(sock, &lsa)
	if err != nil {
		syscall.Close(sock)
		s.logger.Fatal("Fail to bind to address: %s", err.Error())
		panic(err)
	}
	f := os.NewFile(uintptr(sock), "")
	c, err := net.FilePacketConn(f)
	f.Close()
	if err != nil {
		s.logger.Fatal("Cannot start connection: %s", err.Error())
	}
	p := ipv4.NewPacketConn(c)

	itfName, err := net.InterfaceByName(s.itf)
	if err != nil {
		s.logger.Fatal("Unknown interface name %s", s.itf)
	}

	err = p.JoinGroup(itfName, &net.UDPAddr{IP: a})
	if err != nil {
		s.logger.Fatal("Cannot join group %s: %s", a.String(), err.Error())
		panic(err)
	}

	err = p.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	if err != nil {
		s.logger.Fatal("Fail to set socket control message: %s", err.Error())
		panic(err)
	}

	s.conn = p
}

func (s *sockConn) read() []byte {
	buf := make([]byte, 10000)
	n, _, _, err := s.conn.ReadFrom(buf)
	if err != nil {
		s.logger.Fatal("Cannot read from UDP datagram")
		panic(err)
	}
	return buf[:n]
}

func socketConnection(logger common.Log, address string, port string, itf string) *sockConn {
	return &sockConn{logger: logger, address: address, port: port, itf: itf, conn: nil}
}

// Subprocess is used to dump UDP packets instead of gopkts library due to absence of libpcap in my practical use case
// Currently the command always has a timeout
type udpReader struct {
	logger      common.Log
	address     string
	port        string
	itf         string
	timeout     int
	conn        *sockConn
	isRunning   bool
	bufferQueue [][]byte
	bufferSize  int
	udpCount    int
}

func (ur *udpReader) setup() {
	ur.conn = socketConnection(ur.logger, ur.address, ur.port, ur.itf)
}

func (ur *udpReader) startRecv() {
	ur.conn.init()
}

func (ur *udpReader) stopRecv() {
	ur.conn.close()
}

func (ur *udpReader) dataAvailable(unit *common.IOUnit) bool {
	// Overflow
	if len(ur.bufferQueue) > ur.bufferSize {
		ur.logger.Error("Buffer overflow")
		panic("Overflow")
		return true
	}
	if len(ur.bufferQueue) <= 1 {
		udpBuf := ur.conn.read()

		ur.udpCount += 1

		nTsPkt := len(udpBuf) / TS_PKT_SIZE
		// ur.logger.Trace("Fetched %d TS packets from %d-th UDP packet", nTsPkt, ur.udpCount)
		for i := 0; i < nTsPkt; i++ {
			ur.bufferQueue = append(ur.bufferQueue, udpBuf[(i*TS_PKT_SIZE):((i+1)*TS_PKT_SIZE)])
		}
	}

	buf := ur.bufferQueue[0]
	ur.bufferQueue = ur.bufferQueue[1:]

	unit.IoType = 3
	unit.Id = -1
	unit.Buf = buf

	return true
}

func initUdpReader(param *udpInputParam, name string) *udpReader {
	rv := udpReader{}
	rv.conn = nil
	rv.udpCount = 0
	rv.bufferSize = 25000 // 5 Mbps stream * 3 sec buffer / (7 * 188) bps UDP packet = 11399

	rv.logger = common.CreateLogger(name)
	rv.isRunning = false
	rv.bufferQueue = make([][]byte, 0)

	tmp := strings.Split(param.Address, ":")
	rv.address = tmp[0]
	rv.port = tmp[1]
	rv.itf = param.Itf
	rv.timeout = param.Timeout

	return &rv
}
