package common

// Structs to provide a unified interface for communication

type CmUnit interface {
	GetBuf() interface{}
	GetField(name string) interface{}
}

// ====================     Status     ====================

// related status id
type CM_STATUS int

const (
	STATUS_END_ROUTINE  CM_STATUS = 1
	STATUS_CONTROL_DATA CM_STATUS = 2
)

type CmStatusUnit struct {
	status CM_STATUS
	body   interface{}
	id     CM_STATUS // Identify the purpose of the message
}

func MakeStatusUnit(status CM_STATUS, id CM_STATUS, body interface{}) CmStatusUnit {
	return CmStatusUnit{status: status, id: id, body: body}
}

func (unit CmStatusUnit) GetBuf() interface{} {
	return unit.status
}

func (unit CmStatusUnit) GetField(name string) interface{} {
	switch name {
	case "id":
		return unit.id
	case "body":
		return unit.body
	default:
		panic("Unknown field in statusUnit")
	}
}

// ====================     I/O     ====================
type IOUnit struct {
	Buf    interface{}
	IoType int // input: [UNKNOWN, FILE], output: [UNKNOWN, JSON, CSV]
	Id     int // Specify a receiver. If not -1, this is the name of receiver, else hardcoded for now
}

func (unit IOUnit) GetBuf() interface{} {
	return unit.Buf
}

func (unit IOUnit) GetField(name string) interface{} {
	switch name {
	case "type":
		return unit.IoType
	case "id":
		return unit.Id
	default:
		panic("Unknown field in IO unit")
	}
}

// ====================     demuxer     ====================
// A unit for parsing PSI
type PsiUnit struct {
	buf  []byte
	cnt  int
	pid  int
	pusi bool
}

func MakePsiUnit(buf []byte, cnt int, pid int, pusi bool) PsiUnit {
	return PsiUnit{buf: buf, cnt: cnt, pid: pid, pusi: pusi}
}

func (unit PsiUnit) GetBuf() interface{} {
	return unit.buf
}

func (unit PsiUnit) GetField(name string) interface{} {
	switch name {
	case "pid":
		return unit.pid
	case "pusi":
		return unit.pusi
	case "count":
		return unit.cnt
	default:
		panic("Unknown field in PSI unit")
	}
}

// A unit for demuxing PES
type StreamUnit struct {
	buf  []byte
	cnt  int
	pid  int
	pusi bool
	afc  int
}

func MakeStreamUnit(buf []byte, cnt int, pid int, pusi bool, afc int) StreamUnit {
	return StreamUnit{buf: buf, cnt: cnt, pid: pid, pusi: pusi, afc: afc}
}

func (unit StreamUnit) GetBuf() interface{} {
	return unit.buf
}

func (unit StreamUnit) GetField(name string) interface{} {
	switch name {
	case "pid":
		return unit.pid
	case "pusi":
		return unit.pusi
	case "afc":
		return unit.afc
	case "count":
		return unit.cnt
	default:
		panic("Unknown field in stream unit")
	}
}
