package ioUtils

import (
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
	pcap        *pcapFileStruct
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

	ur.pcap = pcapFile(fname, ur.logger)
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
}

func (ur *udpReader) stopRecv() {
	ur.pcap.close()
}

func (ur *udpReader) dataAvailable(unit *common.IOUnit) bool {
	if !ur.isRunning {
		return false
	}
	buf, hasData := ur.pcap.getBufferV2()
	if hasData {
		unit.IoType = 3
		unit.Id = -1
		unit.Buf = buf
		return true
	} else {
		return false
	}
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
	rv.timeout = param.Timeout

	return &rv
}
