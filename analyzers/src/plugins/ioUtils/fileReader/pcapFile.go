package fileReader

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	tio "github.com/tony-507/analyzers/src/common/io"
	"github.com/tony-507/analyzers/src/common/logging"
)

type dataPacketStruct interface {
	parseHeader([]byte, logging.Log)
	setPayload([]byte, logging.Log)
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

func (p *udpPacketStruct) parseHeader(buf []byte, logger logging.Log) {
	r := tio.GetBufferReader(buf)
	p.srcPort = r.ReadBits(16)
	p.dstPort = r.ReadBits(16)
	p.length = r.ReadBits(16)
	p.checksum = r.ReadBits(16)
	p.setPayload(r.GetRemainedBuffer(), logger)
}

func (p *udpPacketStruct) setPayload(buf []byte, logger logging.Log) {
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

func (p *ipv4PacketStruct) parseHeader(buf []byte, logger logging.Log) {
	r := tio.GetBufferReader(buf)
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

func (p *ipv4PacketStruct) setPayload(buf []byte, logger logging.Log) {
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

func (p *ethernetPacketStruct) parseHeader(buf []byte, logger logging.Log) {
	r := tio.GetBufferReader(buf)
	p.dstMAC = strings.Join([]string{r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1)}, ":")
	p.srcMAC = strings.Join([]string{r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1), r.ReadHex(1)}, ":")
	p.etherType = r.ReadBits(16)
	p.setPayload(r.GetRemainedBuffer(), logger)
}

func (p *ethernetPacketStruct) setPayload(buf []byte, logger logging.Log) {
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

func (p *pcapPacketStruct) parseHeader(buf []byte, logger logging.Log) {
	r := tio.GetBufferReader(buf)
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

func (p *pcapPacketStruct) setPayload(buf []byte, logger logging.Log) {
	p.payload = ethernetPacket()
	p.payload.parseHeader(buf, logger)
}

func (p *pcapPacketStruct) getPayload() interface{} {
	return p.payload
}

type pcapFileStruct struct {
	logger        logging.Log
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

func (pcap *pcapFileStruct) parseHeader() error {
	buf := make([]byte, 24)
	pcap.fHandle.Read(buf)
	r := tio.GetBufferReader(buf)
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
		return errors.New(fmt.Sprintf("Unknown first byte of magic number: %d", firstByte))
	}
	r.ReadBits((3 + 2 + 2 + 4 + 4 + 4) * 8)
	if pcap.isBigEndian {
		pcap.linkLayerType = r.ReadBits(32)
	} else {
		pcap.linkLayerType = r.ReadLIBytes(4)
	}
	return nil
}

func (pcap *pcapFileStruct) getBuffer() ([]byte, error) {
	if !pcap.bInit {
		time.Sleep(1 * time.Second)
		// pcap header
		err := pcap.parseHeader()
		if err != nil {
			return []byte{}, err
		}
		pcap.bInit = true
	}

	if len(pcap.bufferQueue) == 0 {
		// Check pcap packet header
		buf, _ := pcap.advanceCursor(16)
		if len(buf) == 0 {
			return buf, nil
		}
		pcapPkt := pcapPacket(pcap.isBigEndian)
		pcapPkt.parseHeader(buf, pcap.logger)

		body, _ := pcap.advanceCursor(pcapPkt.length)
		if len(body) == 0 {
			return body, nil
		}
		pcapPkt.setPayload(body, pcap.logger)

		dataLink, ok := pcapPkt.getPayload().(dataPacketStruct)
		if !ok {
			return buf, errors.New("Fail to get data link packet")
		}

		network, ok := dataLink.getPayload().(dataPacketStruct)
		if !ok {
			return buf, errors.New("Fail to get network packet")
		}

		transport, ok := network.getPayload().(dataPacketStruct)
		if !ok {
			return buf, errors.New("Fail to get transport packet")
		}

		buffer, ok := transport.getPayload().([]byte)
		if !ok {
			return buf, errors.New("Fail to retrieve application payload")
		}

		pcap.bufferQueue = append(pcap.bufferQueue, buffer)
	}

	if len(pcap.bufferQueue) == 0 {
		return []byte{}, errors.New("no output can be fetched")
	} else {
		buf := pcap.bufferQueue[0]
		if len(pcap.bufferQueue) == 1 {
			pcap.bufferQueue = make([][]byte, 0)
		} else {
			pcap.bufferQueue = pcap.bufferQueue[1:]
		}
		return buf, nil
	}
}

// Try to read n bytes. If fail, return false
func (pcap *pcapFileStruct) advanceCursor(n int) ([]byte, error) {
	ok := true
	reason := "unknown"
	buf := make([]byte, n)
	l, err := pcap.fHandle.Read(buf)
	if err == io.EOF {
		reason = "EOF"
		ok = false
	} else if err != nil {
		return buf, err
	}
	if l < n {
		reason = "out of buffer"
		ok = false
	}
	if !ok {
		pcap.logger.Error("FAIL to read %d bytes due to %s, shift back %d bytes", n, reason, l)
		pcap.fHandle.Seek(int64(-l), 1)
		time.Sleep(5 * time.Millisecond)
		return []byte{}, nil
	}
	return buf, nil
}

func pcapFile(handle *os.File, logger logging.Log) fileHandler {
	rv := pcapFileStruct{}
	rv.fHandle = handle
	rv.bufferQueue = make([][]byte, 0)
	rv.logger = logger
	rv.bInit = false

	return &rv
}
