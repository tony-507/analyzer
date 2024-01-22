package tttKernel

import (
	"fmt"
)

type RequestHandler func(string, WORKER_REQUEST, interface{})

// A plugin can send post requests to worker to communicate with other plugins
func Post_request(h RequestHandler, name string, unit CmUnit) {
	if h == nil {
		panic(fmt.Sprintf("Error in sending post request for plugin %s with unit %v", name, unit))
	}
	h(name, POST_REQUEST, unit)
}

func Post_status(h RequestHandler, name string, unit CmUnit) {
	if h == nil {
		panic(fmt.Sprintf("Error in sending status for plugin %s with unit %v", name, unit))
	}
	h(name, STATUS_REQUEST, unit)
}

func Throw_error(h RequestHandler, name string, err error) {
	if h == nil {
		panic(err)
	}
	h(name, ERROR_REQUEST, err)
}

func Listen_msg(h RequestHandler, name string, msgId int) {
	if h == nil {
		panic(fmt.Sprintf("Error in registering status destination for plugin %s with id %d", name, msgId))
	}
	h(name, STATUS_LISTEN_REQUEST, msgId)
}
