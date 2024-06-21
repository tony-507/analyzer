package controller

import (
	"strconv"
	"strings"

	"github.com/tony-507/analyzers/src/tttKernel"
)

type Property struct {
	Name string
	values map[string]Property
}

func NewProperty(name string) Property {
	return Property{
		Name: name,
		values: make(map[string]Property),
	}
}

func (p *Property) AddValue(key string, value Property) {
	p.values[key] = value
}

func (p *Property) toString() string {
	if len(p.values) == 0 {
		_, err := strconv.Atoi(p.Name)
		if err != nil && p.Name != "true" && p.Name != "false" {
			return "\"" + p.Name + "\""
		} else {
			return p.Name
		}
	}

	s := "{"
	fieldArr := make([]string, 0)
	for k, v := range p.values {
		fieldArr = append(fieldArr, "\""+k+"\":"+v.toString())
	}
	s += strings.Join(fieldArr, ",") + "}"
	return s
}

type PluginBuilder struct {
	name       string
	properties map[string]Property
	children   []string
}

func NewPluginBuilder() PluginBuilder {
	return PluginBuilder{
		name:       "",
		properties: make(map[string]Property),
		children:   make([]string, 0),
	}
}

func (pb *PluginBuilder) SetName(name string) {
	pb.name = name
}

func (pb *PluginBuilder) SetProperty(key string, value Property) {
	pb.properties[key] = value
}

func (pb *PluginBuilder) AddChild(child string) {
	pb.children = append(pb.children, child)
}

func (pb *PluginBuilder) Build() tttKernel.OverallParams {
	return tttKernel.ConstructOverallParam(pb.name, pb.getPropertyString(), pb.children)
}

func (pb *PluginBuilder) getPropertyString() string {
	s := "{"
	fieldArr := make([]string, 0)
	for k, v := range pb.properties {
		fieldArr = append(fieldArr, "\""+k+"\":"+v.toString())
	}
	s += strings.Join(fieldArr, ",") + "}"
	return s
}
