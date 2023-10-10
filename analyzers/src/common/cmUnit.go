package common

// Structs to provide a unified interface for communication

/*
 * CmUnit:       The basic interface for different units
 * CmStatusUnit: The unit that stores status information
 * IOUnit:       The basic unit containing data from one plugin to another
 * ReqUnit:      The unit that contains requests to worker
 */

type CmUnit interface {
	GetBuf() CmBuf
	GetField(name string) interface{}
}

func GetBytesInBuf(unit CmUnit) []byte {
	return unit.GetBuf().GetBuf()
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

func MakeStatusUnit(id int, body CmBuf) *CmStatusUnit {
	rv := CmStatusUnit{id: id, body: body}
	return &rv
}

// Unused
func (unit *CmStatusUnit) GetBuf() CmBuf {
	return unit.body
}

func (unit *CmStatusUnit) GetField(name string) interface{} {
	switch name {
	case "id":
		return unit.id
	default:
		panic("Unknown field in statusUnit")
	}
}

// ====================     I/O     ====================
type IOUnit struct {
	Buf    CmBuf
	IoType int // input: [UNKNOWN, FILE], output: [UNKNOWN, JSON, CSV]
	Id     int // Specify a receiver. If not -1, this is the name of receiver, else hardcoded for now
}

func MakeIOUnit(buf CmBuf, ioType int, id int) *IOUnit {
	rv := IOUnit{Buf: buf, IoType: ioType, Id: id}
	return &rv
}

func (unit *IOUnit) GetBuf() CmBuf {
	return unit.Buf
}

func (unit *IOUnit) GetField(name string) interface{} {
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
	STATUS_LISTEN_REQUEST WORKER_REQUEST = 6  // Request to register a status message destination
	STATUS_REQUEST        WORKER_REQUEST = 7  // Send a status
	ERROR_REQUEST         WORKER_REQUEST = 11 // Throw error
)

type reqUnit struct {
	buf     CmBuf
	reqType WORKER_REQUEST
}

func (unit *reqUnit) GetBuf() CmBuf {
	return unit.buf
}

func (unit *reqUnit) GetField(name string) interface{} {
	switch name {
	case "reqType":
		return unit.reqType
	default:
		panic("Unknown field for request unit")
	}
}

func MakeReqUnit(name string, reqType WORKER_REQUEST) *reqUnit {
	buf := MakeSimpleBuf([]byte{})
	rv := reqUnit{buf: buf, reqType: reqType}
	return &rv
}
