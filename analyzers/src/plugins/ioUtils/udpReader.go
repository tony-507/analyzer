package ioUtils

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tony-507/analyzers/src/common"
)

// Subprocess is used to dump UDP packets instead of gopkts library due to absence of libpcap in my practical use case
// Currently the command always has a timeout
type udpReader struct {
	logger      common.Log
	address     string
	port        string
	itf         string
	timeout     int
	fHandle     *os.File
	cmd         *exec.Cmd
	isRunning   bool
	bufferQueue [][]byte
}

func (ur *udpReader) setup() {
	fname := "temp.pcap"
	cmdStr := ""
	if ur.timeout > 0 {
		cmdStr += "timeout  --preserve-status " + strconv.Itoa(ur.timeout) + " "
	}
	cmdStr += strings.Join(
		[]string{"tcpdump", "-i", ur.itf, "'host", ur.address, "and port", ur.port, "'", "-w", fname},
		" ")
	ur.logger.Info("UDP reader command: %s", cmdStr)
	ur.cmd = exec.Command("sh", "-c", cmdStr)

	fHandle, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	ur.fHandle = fHandle
}

func (ur *udpReader) startRecv() {
	go func() {
		ur.isRunning = true
		ur.logger.Info("UDP reader starts receiving data now")
		err := ur.cmd.Run()
		if err != nil {
			panic(err)
		}
		ur.isRunning = false
		ur.logger.Info("UDP reader exits peacefully")
	}()
	// Give some time for tcpdump
	time.Sleep(1 * time.Second)

	// Pcap header
	buf := make([]byte, 24)
	ur.fHandle.Read(buf)
	ur.logger.Info("Pcap header: %v", buf)
}

func (ur *udpReader) stopRecv() {
	ur.fHandle.Close()
}

func (ur *udpReader) dataAvailable(unit *common.IOUnit) bool {
	if !ur.isRunning {
		return false
	}
	hasData := true
	if len(ur.bufferQueue) == 0 {
		time.Sleep(5 * time.Millisecond)
		// Unused packet data
		ur.fHandle.Read(make([]byte, 8))

		// Get packet length
		lenBuf := make([]byte, 4)
		n, err := ur.fHandle.Read(lenBuf)
		if err == io.EOF {
			hasData = false
		} else {
			check(err)
		}
		if n < 4 {
			hasData = false
			ur.fHandle.Seek(int64(-n), 1)
		}

		if hasData {
			// Unused packet data
			ur.fHandle.Read(make([]byte, 4))

			// Physical layer is not needed
			ur.handleLinkLayer()
			ur.handleNetworkLayer()
			pktLen := ur.handleTransportLayer()

			// Read packet payload
			numTsPkt := pktLen / TS_PKT_SIZE
			for i := 0; i < numTsPkt; i++ {
				buf := make([]byte, TS_PKT_SIZE)
				n, err = ur.fHandle.Read(buf)
				if err == io.EOF {
					hasData = false
				} else {
					check(err)
				}
				if n < TS_PKT_SIZE {
					hasData = false
				}
				ur.bufferQueue = append(ur.bufferQueue, buf)
			}
		}
	}

	if hasData {
		unit.IoType = 3
		unit.Id = -1
		unit.Buf = ur.bufferQueue[0]

		if len(ur.bufferQueue) == 1 {
			ur.bufferQueue = make([][]byte, 0)
		} else {
			ur.bufferQueue = ur.bufferQueue[1:]
		}
	}
	return true
}

func (ur *udpReader) handleLinkLayer() {
	// TODO: Assume Ethernet frame: 6 byte dst + 6 byte src + 2 byte type
	ur.fHandle.Read(make([]byte, 14))
}

func (ur *udpReader) handleNetworkLayer() {
	// TODO: Assume IPv4 for now
	firstByte := make([]byte, 1)
	ur.fHandle.Read(firstByte)
	version := (firstByte[0] & 0xf0) / 16
	if version != 4 {
		ur.logger.Error("Internet protocol version is not 4 but %d", version)
	}
	headerLen := firstByte[0] & 0x0f
	ur.fHandle.Read(make([]byte, headerLen*4-1))
}

// Return payload size
func (ur *udpReader) handleTransportLayer() int {
	//TODO: Assume UDP for now: 2 byte src port, 2 byte dst port, 2 byte length, 2 byte checksum
	ur.fHandle.Read(make([]byte, 4))
	lenBuf := make([]byte, 2)
	ur.fHandle.Read(lenBuf)
	ur.fHandle.Read(make([]byte, 2))
	return int(lenBuf[0])*256 + int(lenBuf[1]) - 8
}

func initUdpReader(param *udpInputParam, name string) *udpReader {
	rv := udpReader{}
	rv.logger = common.CreateLogger(name)
	rv.isRunning = false
	rv.bufferQueue = make([][]byte, 0)

	tmp := strings.Split(param.Address, ":")
	rv.address = tmp[0]
	rv.port = tmp[1]
	rv.itf = param.Itf
	rv.timeout = param.timeout

	return &rv
}
