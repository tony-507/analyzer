package common

import (
	"strconv"
)

// General buffer unit design

type CmBuf interface {
	ToString() string         // Return data as byte string
	GetFieldAsString() string // Return name of the fields as byte string
	GetField(name string) interface{}
}

// Buffer unit carrying PSI packet timing data
type PsiBuf struct {
	pktCnt int
	pid    int
	pcr    int
}

func (pb *PsiBuf) ToString() string {
	rv := ""
	rv += strconv.Itoa(pb.pktCnt)
	rv += ","
	rv += strconv.Itoa(pb.pid)
	rv += ","
	rv += strconv.Itoa(pb.pcr / 300)
	rv += "\n"
	return rv
}

func (pb *PsiBuf) GetFieldAsString() string {
	return "pktCnt,pid,pcr\n"
}

func (pb *PsiBuf) GetField(name string) interface{} {
	switch name {
	case "pktCnt":
		return pb.pktCnt
	}
	return -1
}

func (pb *PsiBuf) SetPcr(pcr int) {
	pb.pcr = pcr
}

func MakePsiBuf(pktCnt int, pid int) PsiBuf {
	return PsiBuf{pktCnt: pktCnt, pid: pid}
}

// Buffer unit carrying PES packet data

type PesBuf struct {
	pktCnt  int
	progNum int
	size    int
	pts     int
	dts     int
	pcr     int
	delay   int
}

func (pb *PesBuf) ToString() string {
	rv := ""
	rv += strconv.Itoa(pb.pktCnt)
	rv += ","
	rv += strconv.Itoa(pb.size)
	rv += ","
	rv += strconv.Itoa(pb.pcr / 300)
	rv += ","
	rv += strconv.Itoa(pb.pts)
	rv += ","
	rv += strconv.Itoa(pb.dts)
	rv += ","
	rv += strconv.Itoa(pb.delay)
	rv += "\n"

	return rv
}

func (pb *PesBuf) GetFieldAsString() string {
	return "pktCnt,size,pcr,pts,dts,delay\n"
}

func (pb *PesBuf) GetField(name string) interface{} {
	switch name {
	case "pktCnt":
		return pb.pktCnt
	case "progNum":
		return pb.progNum
	}
	return -1
}

func (pb *PesBuf) SetPcr(pcr int) {
	pb.pcr = pcr
	if pcr != -1 {
		pb.delay = pb.dts - pb.pcr/300
	} else {
		pb.delay = -1
	}
}

func MakePesBuf(pktCnt int, progNum int, size int, pts int, dts int) PesBuf {
	return PesBuf{pktCnt: pktCnt, progNum: progNum, size: size, pts: pts, dts: dts}
}
