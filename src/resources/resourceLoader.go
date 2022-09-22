package resources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Resource struct {
	StreamType []string `json:"StreamType"`
}

type ResourceLoader struct {
	resource Resource
}

func (r *ResourceLoader) Query(path string, key interface{}) string {
	sPath := strings.Split(path, "/")
	switch sPath[0] {
	case "streamType":
		typeNum, ok := key.(int)
		if !ok {
			panic("Wrong query format for streamType")
		}
		return r.resource.StreamType[typeNum-1]
	}
	return ""
}

func CreateResourceLoader() ResourceLoader {
	f, err := os.Open("resources/app.json")
	if err != nil {
		fmt.Println("ERROR: Resources not found")
		os.Exit(1)
	}
	buf, _ := ioutil.ReadAll(f)

	resource := Resource{}
	json.Unmarshal(buf, &resource)

	return ResourceLoader{resource: resource}
}
