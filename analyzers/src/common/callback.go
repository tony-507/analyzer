package common

import "fmt"

type RequestHandler func(string, WORKER_REQUEST, interface{})

// A plugin can send post requests to worker to communicate with other plugins
func Post_request(h RequestHandler, name string, unit CmUnit) {
	// For debug
	if h == nil {
		msg := fmt.Sprintf("Error in sending post request for plugin %s with unit %v", name, unit)
		panic(msg)
	}
	h(name, POST_REQUEST, unit)
}

func Throw_error(h RequestHandler, name string, err error) {
	// For debug
	if h == nil {
		panic(err)
	}
	h(name, ERROR_REQUEST, err)
}
