package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclarePlugin(t *testing.T) {
	script := "test = #Dummy_1;"
	input := []string{}
	ctrl := newController()

	ctrl.parser.buildParams(script, input, -1)

	assert.Equal(t, "test", ctrl.parser.variables[0].name, "Name of plugin is not test")
	assert.Equal(t, "Dummy_1", ctrl.parser.variables[0].value, "Value of plugin is not Dummy")
}

func TestDeclareVarInScript(t *testing.T) {
	script := "x = $x; x.a = $yes;"
	input := []string{"--yes", "bye", "-x", "hi"}
	ctrl := newController()

	ctrl.parser.buildParams(script, input, -1)

	assert.Equal(t, "x", ctrl.parser.variables[0].name, "Name of x is not x")
	assert.Equal(t, "hi", ctrl.parser.variables[0].value, "Value of x is not hi")
	assert.Equal(t, "a", ctrl.parser.variables[0].attributes[0].name, "Name of x.a is not a")
	assert.Equal(t, "bye", ctrl.parser.variables[0].attributes[0].value, "Value of x.a is not bye")
}

func TestSetAlias(t *testing.T) {
	script := "alias(test, x); x = $x;"
	input := []string{"--test", "hi"}
	ctrl := newController()

	ctrl.parser.buildParams(script, input, -1)

	assert.Equal(t, "hi", ctrl.parser.variables[0].value, "Value of x is not hi")
}

func TestRunNestedConditional(t *testing.T) {
	script := "if $x; x = $x; if $y; x = $y; end; end;"
	input := []string{"-x", "hi", "-y", "bye"}
	ctrl := newController()

	ctrl.parser.buildParams(script, input, -1)

	assert.Equal(t, "bye", ctrl.parser.variables[0].value, "Nested conditional fails")
}

func TestRunPartialNestedConditional(t *testing.T) {
	script := "if $x; x = $x; if $y; x = $y; end; end;"
	input := []string{"-x", "hi"}
	ctrl := newController()

	ctrl.parser.buildParams(script, input, -1)

	assert.Equal(t, "hi", ctrl.parser.variables[0].value, "Nested conditional fails")
}

func TestGetEmptyAttributeString(t *testing.T) {
	v := scriptVar{name: "dummy", varType: _VAR_PLUGIN, value: "dummy_1", attributes: make([]*scriptVar, 0)}
	s := v.getAttributeStr()

	assert.Equal(t, "{}", s, "Fail to get correct attribute string for a plugin with empty parameter")
}

func TestGetRecursiveAttributeString(t *testing.T) {
	v := scriptVar{name: "dummy", varType: _VAR_PLUGIN, value: "dummy_1", attributes: make([]*scriptVar, 0)}
	x := scriptVar{name: "x", varType: _VAR_VALUE, value: "", attributes: make([]*scriptVar, 0)}
	y := scriptVar{name: "y", varType: _VAR_VALUE, value: "3", attributes: make([]*scriptVar, 0)}
	a := scriptVar{name: "a", varType: _VAR_VALUE, value: "abc", attributes: make([]*scriptVar, 0)}

	v.attributes = append(v.attributes, &x)
	v.attributes = append(v.attributes, &y)
	v.attributes[0].attributes = append(v.attributes[0].attributes, &a)

	s := v.getAttributeStr()

	assert.Equal(t, "{\"x\":{\"a\":\"abc\"},\"y\":3}", s, "Fail to get correct attribute string for a plugin with recursive parameters")
}
