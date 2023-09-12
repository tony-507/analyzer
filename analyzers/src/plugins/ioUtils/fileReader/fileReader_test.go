package fileReader

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
)

func getOutputDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

var ASSET_DIR = getOutputDir() + "/../../../../test/resources/assets/"

func TestReadPcapPacket(t *testing.T) {
	// Build data
	pcapData := []byte{0x3d, 0x17, 0x9c, 0x63, 0x14, 0x08, 0x08, 0x00, 0x4e, 0x05, 0x00, 0x00, 0x4e, 0x05, 0x00, 0x00}
	ethernetData := []byte{0x01, 0x00, 0x5e, 0x01, 0x01, 0x01, 0x00, 0x1e, 0x67, 0xd1, 0x1c, 0xe4, 0x08, 0x00}
	ipv4Data := []byte{0x45, 0x00, 0x00, 0x1e, 0xf7, 0x5a, 0x00, 0x00, 0x40, 0x11, 0xe0, 0x30, 0xac, 0x12, 0x0f, 0x0d, 0xe2, 0x01, 0x01, 0x01}
	udpData := []byte{0xb4, 0x46, 0x30, 0x22, 0x00, 0x0a, 0xa3, 0x5f, 0x01, 0x02}
	data := []byte{}
	data = append(data, pcapData...)
	data = append(data, ethernetData...)
	data = append(data, ipv4Data...)
	data = append(data, udpData...)

	logger := common.CreateLogger("Dummy")

	pcap := pcapPacket(false)
	pcap.parseHeader(data, logger)
	assert.Equal(t, 1671173949, pcap.sec)
	assert.Equal(t, 526356, pcap.msec)
	assert.Equal(t, 1358, pcap.length)
	assert.Equal(t, 1358, pcap.origLength)
	pcap.setPayload(data[16:], logger)

	eth, ok := pcap.getPayload().(*ethernetPacketStruct)
	if !ok {
		panic("Data link layer is not Ethernet")
	}
	assert.Equal(t, "01:00:5e:01:01:01", eth.dstMAC)
	assert.Equal(t, "00:1e:67:d1:1c:e4", eth.srcMAC)
	assert.Equal(t, 0x0800, eth.etherType)

	ip, ok := eth.getPayload().(*ipv4PacketStruct)
	if !ok {
		panic("Network layer is not IPv4")
	}
	assert.Equal(t, 20, ip.headerLength)
	assert.Equal(t, 30, ip.length)
	assert.Equal(t, "172.18.15.13", ip.srcIp)
	assert.Equal(t, "226.1.1.1", ip.dstIp)

	udp, ok := ip.getPayload().(*udpPacketStruct)
	if !ok {
		panic("Transport layer is not UDP packet")
	}
	assert.Equal(t, 46150, udp.srcPort)
	assert.Equal(t, 12322, udp.dstPort)
	assert.Equal(t, 10, udp.length)
	assert.Equal(t, []byte{0x01, 0x02}, udp.payload)
}

func TestReadPcapFile(t *testing.T) {
	fname := ASSET_DIR + "adSmart.pcap"
	logger := common.CreateLogger("Dummy")
	handle, err := os.Open(fname)
	pcap, err := pcapFile(handle, logger)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 50; i++ {
		buf, err := pcap.getBuffer()
		if err != nil {
			panic(err)
		}
		assert.Equal(t, uint8(0x47), buf[0], "TS sync byte not match")
	}
	assert.Equal(t, 0x01, pcap.linkLayerType, "Link layer type is not Ethernet")
}
