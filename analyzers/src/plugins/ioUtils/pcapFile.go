package ioUtils

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tony-507/analyzers/src/common"
)

type dataPacketStruct interface {
	parseHeader([]byte, common.Log)
	setPayload([]byte, common.Log)
	getPayload() interface{}
}

type udpPacketStruct struct {
	srcPort  int
	dstPort  int
	length   int
	checksum int
	payload  []byte
}

func udpPacket() *udpPacketStruct {
	return &udpPacketStruct{}
}

func (p *udpPacketStruct) parseHeader(buf []byte, logger common.Log) {
	r := common.GetBufferReader(buf)
	p.srcPort = r.ReadBits(16)
	p.dstPort = r.ReadBits(16)
	p.length = r.ReadBits(16)
	p.checksum = r.ReadBits(16)
	p.setPayload(r.GetRemainedBuffer(), logger)
}

func (p *udpPacketStruct) setPayload(buf []byte, logger common.Log) {
	p.payload = buf
}

func (p *udpPacketStruct) getPayload() interface{} {
	return p.payload
}

type ipv4PacketStruct struct {
	headerLength int
	length       int
	checksum     int
	srcIp        string
	dstIp        string
	payload      dataPacketStruct
}

func ipv4Packet() *ipv4PacketStruct {
	return &ipv4PacketStruct{}
}

func (p *ipv4PacketStruct) parseHeader(buf []byte, logger common.Log) {
	r := common.GetBufferReader(buf)
	version := r.ReadBits(4)
	if version != 4 {
		logger.Error("Internet protocol version is not 4 but %d", version)
	}
	p.headerLength = r.ReadBits(4) * 4
	r.ReadBits(8)
	p.length = r.ReadBits(16)
	r.ReadBits(48)
	p.checksum = r.ReadBits(16)
	p.srcIp = strings.Join([]string{
		strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8))}, ".")
	p.dstIp = strings.Join([]string{
		strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8)), strconv.Itoa(r.ReadBits(8))}, ".")
	r.ReadBits(p.headerLength*8 - 160)
	p.setPayload(r.GetRemainedBuffer(), logger)
}

func (p *ipv4PacketStruct) setPayload(buf []byte, logger common.Log) {
	p.payload = udpPacket()
	p.payload.parseHeader(buf, logger)
}

func (p *ipv4PacketStruct) getPayload() interface{} {
	return p.payload
}

type ethernetPacketStruct struct {
	dstMAC    string
	srcMAC    string
	etherType int
	payload   dataPacketStruct
}

func ethernetPacket() *ethernetPacketStruct {
	return &ethernetPacketStruct{}
}

func (p *ethernetPacketStruct) parseHeader(buf []byte, logger common.Log) {
	r := common.GetBufferReader(buf)
	p.dstMAC = strings.Join([]string{r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1)}, ":")
	p.srcMAC = strings.Join([]string{r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1)}, ":")
	p.etherType = r.ReadBits(16)
	p.setPayload(r.GetRemainedBuffer(), logger)
}

func (p *ethernetPacketStruct) setPayload(buf []byte, logger common.Log) {
	if p.etherType == 0x0800 {
		p.payload = ipv4Packet()
	} else {
		logger.Fatal("EtherType %d is not supported", p.etherType)
	}
	p.payload.parseHeader(buf, logger)
}

func (p *ethernetPacketStruct) getPayload() interface{} {
	return p.payload
}

type pcapPacketStruct struct {
	sec         int
	msec        int
	length      int
	origLength  int
	isBigEndian bool
	payload     dataPacketStruct
}

func pcapPacket(isBigEndian bool) *pcapPacketStruct {
	return &pcapPacketStruct{isBigEndian: isBigEndian}
}

func (p *pcapPacketStruct) parseHeader(buf []byte, logger common.Log) {
	r := common.GetBufferReader(buf)
	if p.isBigEndian {
		p.sec = r.ReadBits(32)
		p.msec = r.ReadBits(32)
		p.length = r.ReadBits(32)
		p.origLength = r.ReadBits(32)
	} else {
		p.sec = r.ReadLIBytes(4)
		p.msec = r.ReadLIBytes(4)
		p.length = r.ReadLIBytes(4)
		p.origLength = r.ReadLIBytes(4)
	}
}

func (p *pcapPacketStruct) setPayload(buf []byte, logger common.Log) {
	p.payload = ethernetPacket()
	p.payload.parseHeader(buf, logger)
}

func (p *pcapPacketStruct) getPayload() interface{} {
	return p.payload
}

type pcapFileStruct struct {
	logger        common.Log
	fHandle       *os.File
	isBigEndian   bool
	useNanoSec    bool
	linkLayerType int
	pktCnt        int
	bufferQueue   [][]byte
	bInit         bool
}

func (pcap *pcapFileStruct) close() {
	pcap.fHandle.Close()
}

func (pcap *pcapFileStruct) parseHeader() {
	buf := make([]byte, 24)
	pcap.fHandle.Read(buf)
	r := common.GetBufferReader(buf)
	// Determine structure from magic number
	firstByte := r.ReadBits(8)
	switch firstByte {
	case 0xd4:
		pcap.isBigEndian = false
		pcap.useNanoSec = false
	case 0x4d:
		pcap.isBigEndian = false
		pcap.useNanoSec = true
	case 0xa1:
		pcap.isBigEndian = true
		pcap.useNanoSec = false
	case 0x1a:
		pcap.isBigEndian = true
		pcap.useNanoSec = true
	default:
		panic(fmt.Sprintf("Unknown first byte of magic number: %d", firstByte))
	}
	r.ReadBits((3 + 2 + 2 + 4 + 4 + 4) * 8)
	if pcap.isBigEndian {
		pcap.linkLayerType = r.ReadBits(32)
	} else {
		pcap.linkLayerType = r.ReadLIBytes(4)
	}
}

func (pcap *pcapFileStruct) getBufferV2() ([]byte, bool) {
	if !pcap.bInit {
		time.Sleep(1 * time.Second)
		// pcap header
		pcap.parseHeader()
		pcap.bInit = true
	}

	if len(pcap.bufferQueue) == 0 {
		// Check pcap packet header
		buf, hasData := pcap.advanceCursor(16)
		if !hasData {
			return buf, hasData
		}
		pcapPkt := pcapPacket(pcap.isBigEndian)
		pcapPkt.parseHeader(buf, pcap.logger)
		body, hasData := pcap.advanceCursor(pcapPkt.length)
		if hasData {
			pcapPkt.setPayload(body, pcap.logger)
		}
		dataLink, ok := pcapPkt.getPayload().(dataPacketStruct)
		if !ok {
			panic("Fail to get data link packet")
		}
		network, ok := dataLink.getPayload().(dataPacketStruct)
		if !ok {
			panic("Fail to get network packet")
		}
		transport, ok := network.getPayload().(dataPacketStruct)
		if !ok {
			panic("Fail to get transport packet")
		}
		buffer, ok := transport.getPayload().([]byte)
		if !ok {
			panic("Fail to retrieve application payload")
		}
		numTsPkt := len(buffer) / TS_PKT_SIZE
		for i := 0; i < numTsPkt; i++ {
			pcap.bufferQueue = append(pcap.bufferQueue, buffer[(i*TS_PKT_SIZE):((i+1)*TS_PKT_SIZE)])
		}
	}

	if len(pcap.bufferQueue) == 0 {
		return []byte{}, false
	} else {
		buf := pcap.bufferQueue[0]
		if len(pcap.bufferQueue) == 1 {
			pcap.bufferQueue = make([][]byte, 0)
		} else {
			pcap.bufferQueue = pcap.bufferQueue[1:]
		}
		return buf, true
	}
}

// Try to read n bytes. If fail, return false
func (pcap *pcapFileStruct) advanceCursor(n int) ([]byte, bool) {
	ok := true
	reason := "unknown"
	buf := make([]byte, n)
	l, err := pcap.fHandle.Read(buf)
	if err == io.EOF {
		reason = "EOF"
		ok = false
	} else if err != nil {
		panic(err)
	}
	if l < n {
		reason = "out of buffer"
		ok = false
	}
	if !ok {
		pcap.logger.Error("FAIL to read %d bytes due to %s, shift back %d bytes", n, reason, l)
		pcap.logger.Info("Read buffer: %v", buf[:l])
		pcap.fHandle.Seek(int64(-l), 1)
		newBuf := make([]byte, n)
		nl, _ := pcap.fHandle.Read(newBuf)
		pcap.logger.Info("Shifted buffer of size %d: %v", nl, newBuf[:nl])
	}
	return buf, ok
}

func (pcap *pcapFileStruct) handleLinkLayer() (int, bool) {
	// TODO: Assume Ethernet frame: 6 byte dst + 6 byte src + 2 byte type
	_, ok := pcap.advanceCursor(14)
	return -1, ok
}

func (pcap *pcapFileStruct) handleNetworkLayer() (int, bool) {
	// TODO: Assume IPv4 for now
	firstByte, ok := pcap.advanceCursor(1)
	if !ok {
		return -1, ok
	}

	version := (firstByte[0] & 0xf0) / 16
	if version != 4 {
		pcap.logger.Error("Internet protocol version is not 4 but %d", version)
	}

	headerLen := firstByte[0] & 0x0f
	if headerLen != 0 {
		_, ok = pcap.advanceCursor(int(headerLen)*4 - 1)
	}
	return -1, ok
}

// Return payload size
func (pcap *pcapFileStruct) handleTransportLayer() (int, bool) {
	//TODO: Assume UDP for now: 2 byte src port, 2 byte dst port, 2 byte length, 2 byte checksum
	_, ok := pcap.advanceCursor(4)
	if !ok {
		return -1, ok
	}
	lenBuf, ok := pcap.advanceCursor(2)
	if !ok {
		return -1, ok
	}
	_, ok = pcap.advanceCursor(2)
	return (int(lenBuf[0])*256 + int(lenBuf[1]) - 8), ok
}

func (pcap *pcapFileStruct) getBuffer() ([]byte, bool) {
	if !pcap.bInit {
		time.Sleep(1 * time.Second)
		// pcap header
		pcap.advanceCursor(24)
		pcap.bInit = true
	}
	hasData := true
	rv := make([]byte, TS_PKT_SIZE)
	if len(pcap.bufferQueue) == 0 {
		time.Sleep(5 * time.Millisecond)
		pcap.pktCnt += 1
		pcap.logger.Info("Packet #%d", pcap.pktCnt)
		// Unused packet data
		_, hasData = pcap.advanceCursor(8)

		if hasData {
			// Get packet length
			pcapPktLenBuf, canReadLen := pcap.advanceCursor(4)
			if canReadLen {
				_, hasData = pcap.advanceCursor(int(pcapPktLenBuf[0]))
				if !hasData {
					return rv, false
				}
			} else {
				hasData = false
			}
		}

		if hasData {
			// Unused packet data
			_, hasData = pcap.advanceCursor(4)

		}

		// Physical layer is not needed
		if hasData {
			_, hasData = pcap.handleLinkLayer()
		}

		if hasData {
			_, hasData = pcap.handleNetworkLayer()
		}

		if hasData {
			var pktLen int
			pktLen, hasData = pcap.handleTransportLayer()

			if hasData {
				// Read packet payload
				numTsPkt := pktLen / TS_PKT_SIZE
				for i := 0; i < numTsPkt; i++ {
					buf := make([]byte, TS_PKT_SIZE)
					buf, hasData = pcap.advanceCursor(TS_PKT_SIZE)
					if hasData {
						pcap.bufferQueue = append(pcap.bufferQueue, buf)
					} else {
						break
					}
				}
			}
		}
	}

	if hasData {
		if len(pcap.bufferQueue) == 0 {
			pcap.logger.Info("Empty buffer queue is unexpected!")
		}
		rv = pcap.bufferQueue[0]

		if len(pcap.bufferQueue) == 1 {
			pcap.bufferQueue = make([][]byte, 0)
		} else {
			pcap.bufferQueue = pcap.bufferQueue[1:]
		}
	}
	return rv, hasData
}

func pcapFile(fname string, logger common.Log) *pcapFileStruct {
	rv := pcapFileStruct{}
	handle, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	rv.fHandle = handle
	rv.bufferQueue = make([][]byte, 0)
	rv.logger = logger
	rv.bInit = false

	return &rv
}
