package common

// Structs to provide a unified interface for communication

/*
 * CmUnit:       The basic interface for different units
 * CmStatusUnit: The unit that stores status information
 * IOUnit:       The basic unit containing data from one plugin to another
 * ReqUnit:      The unit that contains requests to worker
 */

type CmUnit interface {
	GetBuf() interface{}
	GetField(name string) interface{}
}

// ====================     Status     ====================

// related status id
type CM_STATUS int

const (
	STATUS_END_ROUTINE CM_STATUS = 1
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
	IoType int // input: [UNKNOWN, FILE], output: [UNKNOWN, JSON, CSV, RAW]
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

// ====================     worker     ====================
type WORKER_REQUEST int

const (
	POST_REQUEST     WORKER_REQUEST = 0  // Request type < 10 are all post requests
	FETCH_REQUEST    WORKER_REQUEST = 1  // Ready for fetch
	DELIVER_REQUEST  WORKER_REQUEST = 2  // Request for input
	EOS_REQUEST      WORKER_REQUEST = 3  // Root has nothing more to do, please stop
	RESOURCE_REQUEST WORKER_REQUEST = 4  // Request for querying resourceLoader. buf should be ["path", "key"]
	ERROR_REQUEST    WORKER_REQUEST = 11 // Throw error
)

type ReqUnit struct {
	buf     interface{}
	reqType WORKER_REQUEST
}

func (unit ReqUnit) GetBuf() interface{} {
	return unit.buf
}

func (unit ReqUnit) GetField(name string) interface{} {
	switch name {
	case "reqType":
		return unit.reqType
	default:
		panic("Unknown field for request unit")
	}
}

func MakeReqUnit(buf interface{}, reqType WORKER_REQUEST) ReqUnit {
	return ReqUnit{buf: buf, reqType: reqType}
}
