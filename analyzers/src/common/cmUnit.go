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

/*
 * CmStatusUnit
 *
 * This is a type that allows communication of plugins independent of the working graph for plugin configuration update.
 * Plugins can listen to a particular status type. No quit option is allowed.
 * Status is assumed to be immutable and has unique id. IDs less than 10 are reserved for common use.
 */
const (
	STATUS_END_ROUTINE int = 1
)

type CmStatusUnit struct {
	body CmBuf
	id   int
}

func MakeStatusUnit(id int, body CmBuf) CmStatusUnit {
	return CmStatusUnit{id: id, body: body}
}

// Unused
func (unit CmStatusUnit) GetBuf() interface{} {
	return unit.body
}

func (unit CmStatusUnit) GetField(name string) interface{} {
	switch name {
	case "id":
		return unit.id
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
	POST_REQUEST          WORKER_REQUEST = 0  // Request type < 10 are all post requests
	FETCH_REQUEST         WORKER_REQUEST = 1  // Ready for fetch
	DELIVER_REQUEST       WORKER_REQUEST = 2  // Request for input
	EOS_REQUEST           WORKER_REQUEST = 3  // Root has nothing more to do, please stop
	RESOURCE_REQUEST      WORKER_REQUEST = 4  // Request for querying resourceLoader. buf should be ["path", "key"]
	STATUS_LISTEN_REQUEST WORKER_REQUEST = 6  // Request to register a status message destination
	STATUS_REQUEST        WORKER_REQUEST = 7  // Send a status
	ERROR_REQUEST         WORKER_REQUEST = 11 // Throw error
)

type reqUnit struct {
	buf     interface{}
	reqType WORKER_REQUEST
}

func (unit reqUnit) GetBuf() interface{} {
	return unit.buf
}

func (unit reqUnit) GetField(name string) interface{} {
	switch name {
	case "reqType":
		return unit.reqType
	default:
		panic("Unknown field for request unit")
	}
}

func MakeReqUnit(buf interface{}, reqType WORKER_REQUEST) reqUnit {
	return reqUnit{buf: buf, reqType: reqType}
}
