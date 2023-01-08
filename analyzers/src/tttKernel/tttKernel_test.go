package tttKernel

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/src/logs"
)

func TestPluginInterfaces(t *testing.T) {
	pg := getPluginByName("Dummy_1")
	pg.impl.SetCallback(func(string, common.WORKER_REQUEST, interface{}) {

	})

	inUnit := common.MakeIOUnit(2, 0, 0)
	pg.impl.DeliverUnit(inUnit)

	unit := pg.impl.FetchUnit()

	rv, isInt := unit.GetBuf().(int)
	if !isInt {
		panic("return value is not an integer")
	}
	if rv != 20 {
		panic(fmt.Sprintf("Expect return value 20, but got %d", rv))
	}
}

func TestSimpleGraph(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root")
	dummy2 := getPluginByName("Dummy_2")
	dummy3 := getPluginByName("Dummy_3")

	nodes := []*graphNode{&dummy1, &dummy2, &dummy3}

	// root -> 2 -> 3
	addPath(&dummy1, []*graphNode{&dummy2})
	addPath(&dummy2, []*graphNode{&dummy3})

	w := getWorker()
	w.setGraph(nodes)

	w.runGraph()

	unit := dummy3.impl.FetchUnit()
	cnt, _ := unit.GetBuf().(int)

	assert.Equal(t, 20001, cnt, "Count should be 20001")
}

func TestGraphMultipleInput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root")
	dummy2 := getPluginByName("Dummy_2")
	dummy3 := getPluginByName("Dummy_3")
	dummy4 := getPluginByName("Dummy_4")

	nodes := []*graphNode{&dummy1, &dummy2, &dummy3, &dummy4}
	//   -> 2
	// 1      -> 4
	//   -> 3

	addPath(&dummy1, []*graphNode{&dummy2, &dummy3})
	addPath(&dummy2, []*graphNode{&dummy4})
	addPath(&dummy3, []*graphNode{&dummy4})

	w := getWorker()
	w.setGraph(nodes)

	w.runGraph()

	unit := dummy4.impl.FetchUnit()
	cnt, _ := unit.GetBuf().(int)

	assert.Equal(t, 40002, cnt, "Count should be 40002")
}

func TestGraphMultipleOutput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root")
	dummy2 := getPluginByName("Dummy_2")
	dummy3 := getPluginByName("Dummy_3")
	dummy4 := getPluginByName("Dummy_4")

	nodeList := []*graphNode{&dummy1, &dummy2, &dummy3, &dummy4}
	//        -> 3
	// 1 -> 2
	//        -> 4
	addPath(&dummy1, []*graphNode{&dummy2})
	addPath(&dummy2, []*graphNode{&dummy3, &dummy4})

	w := getWorker()
	w.setGraph(nodeList)

	w.runGraph()

	unit1 := dummy3.impl.FetchUnit()
	cnt1, _ := unit1.GetBuf().(int)

	assert.Equal(t, 20001, cnt1, "First count should be 20001")

	unit2 := dummy4.impl.FetchUnit()
	cnt2, _ := unit2.GetBuf().(int)

	assert.Equal(t, 20001, cnt2, "Second count should be 20001")
}

func TestGraphBuilding(t *testing.T) {
	// Since graph uses pointers to store plugins, we cannot compare the constructed graph with one built manually
	// As an alternative, we use representative fields to compare the graph
	dummyParam1 := constructOverallParam("Dummy_1", "{}", []string{"Dummy_2"})
	dummyParam2 := constructOverallParam("Dummy_2", "{}", []string{"Dummy_3"})
	dummyParam3 := constructOverallParam("Dummy_3", "{}", []string{})

	builtOutput := buildGraph([]OverallParams{dummyParam1, dummyParam2, dummyParam3})

	// Check each node
	for idx, pg := range builtOutput {
		pgName := "Dummy_" + strconv.Itoa(idx+1)
		if pg.impl.Name() != pgName {
			panic(fmt.Sprintf("Name not match. Expecting %s, but got %s", pgName, pg.impl.Name()))
		}
		assert.Equal(t, "{}", pg.m_parameter)
		if idx != 2 {
			if len(pg.children) != 1 {
				panic(fmt.Sprintf("Output edge count not match. Expecting 1, but got %d", len(pg.children)))
			}
			nextName := "Dummy_" + strconv.Itoa(idx+2)
			if pg.children[0].impl.Name() != nextName {
				panic(fmt.Sprintf("Children edge not match. Expecting %s, but got %s", pg.children[0].impl.Name(), nextName))
			}
		}
	}
}

func TestDeclareVarInScript(t *testing.T) {
	script := "x = $x; x.a = $yes;"
	input := []string{"--yes", "bye", "-x", "hi"}
	ctrl := tttKernel{
		logger:    logs.CreateLogger("Controller"),
		variables: []*scriptVar{},
		edgeMap:   map[string][]string{},
		aliasMap:  map[string]string{},
	}

	ctrl.buildParams(script, input)

	assert.Equal(t, "x", ctrl.variables[0].name, "Name of x is not x")
	assert.Equal(t, "hi", ctrl.variables[0].value, "Value of x is not hi")
	assert.Equal(t, "a", ctrl.variables[0].attributes[0].name, "Name of x.a is not a")
	assert.Equal(t, "bye", ctrl.variables[0].attributes[0].value, "Value of x.a is not bye")
}

func TestSetAlias(t *testing.T) {
	script := "alias(test, x); x = $x;"
	input := []string{"--test", "hi"}
	ctrl := tttKernel{
		logger:    logs.CreateLogger("Controller"),
		variables: []*scriptVar{},
		edgeMap:   map[string][]string{},
		aliasMap:  map[string]string{},
	}

	ctrl.buildParams(script, input)

	assert.Equal(t, "hi", ctrl.variables[0].value, "Value of x is not hi")
}

func TestRunNestedConditional(t *testing.T) {
	script := "if $x; x = $x; if $y; x = $y; end; end;"
	input := []string{"-x", "hi", "-y", "bye"}
	ctrl := tttKernel{
		logger:    logs.CreateLogger("Controller"),
		variables: []*scriptVar{},
		edgeMap:   map[string][]string{},
		aliasMap:  map[string]string{},
	}

	ctrl.buildParams(script, input)

	assert.Equal(t, "bye", ctrl.variables[0].value, "Nested conditional fails")
}

func TestRunPartialNestedConditional(t *testing.T) {
	script := "if $x; x = $x; if $y; x = $y; end; end;"
	input := []string{"-x", "hi"}
	ctrl := tttKernel{
		logger:    logs.CreateLogger("Controller"),
		variables: []*scriptVar{},
		edgeMap:   map[string][]string{},
		aliasMap:  map[string]string{},
	}

	ctrl.buildParams(script, input)

	assert.Equal(t, "hi", ctrl.variables[0].value, "Nested conditional fails")
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
