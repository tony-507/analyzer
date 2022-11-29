package common

import "fmt"

// A callback coordinates work on different plugins

// A plugin can send post requests to worker to communicate with other plugins
type PostRequestHandler func(string, CmUnit)

func Post_request(h PostRequestHandler, name string, unit CmUnit) {
	// For debug
	if h == nil {
		msg := fmt.Sprintf("Error in sending post request for plugin %s with unit %v", name, unit)
		panic(msg)
	}
	h(name, unit)
}
