package tttKernel

import (
	"fmt"

	"github.com/tony-507/analyzers/src/common"
)

type RequestHandler func(string, common.WORKER_REQUEST, interface{})

// A plugin can send post requests to worker to communicate with other plugins
func Post_request(h RequestHandler, name string, unit common.CmUnit) {
	if h == nil {
		panic(fmt.Sprintf("Error in sending post request for plugin %s with unit %v", name, unit))
	}
	h(name, common.POST_REQUEST, unit)
}

func Post_status(h RequestHandler, name string, unit common.CmUnit) {
	if h == nil {
		panic(fmt.Sprintf("Error in sending status for plugin %s with unit %v", name, unit))
	}
	h(name, common.STATUS_REQUEST, unit)
}

func Throw_error(h RequestHandler, name string, err error) {
	if h == nil {
		panic(err)
	}
	h(name, common.ERROR_REQUEST, err)
}

func Listen_msg(h RequestHandler, name string, msgId int) {
	if h == nil {
		panic(fmt.Sprintf("Error in registering status destination for plugin %s with id %d", name, msgId))
	}
	h(name, common.STATUS_LISTEN_REQUEST, msgId)
}
