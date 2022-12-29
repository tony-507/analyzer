package worker

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tony-507/analyzers/src/common"
)

func TestPluginInterfaces(t *testing.T) {
	pg := GetPluginByName("Dummy_1")
	assert.Equal(t, "*worker.DummyPlugin", fmt.Sprintf("%T", pg.Work), "Wrong type")

	inUnit := common.IOUnit{Buf: 2, IoType: 0, Id: 0}
	pg.deliverUnit(inUnit)

	unit := pg.fetchUnit()

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
	dummy1 := GetPluginByName("Dummy_root")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")

	// 1 -> 2 -> 3
	graph := GetEmptyGraph()
	graph.AddNode(&dummy1)
	graph.AddNode(&dummy2)
	graph.AddNode(&dummy3)

	AddPath(&dummy1, []*Plugin{&dummy2})
	AddPath(&dummy2, []*Plugin{&dummy3})

	w := GetWorker()
	w.SetGraph(graph)

	w.RunGraph()

	unit := dummy3.fetchUnit()
	cnt, _ := unit.GetBuf().(int)

	assert.Equal(t, 20001, cnt, "Count should be 20001")
}

func TestGraphMultipleInput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := GetPluginByName("Dummy_root")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")
	dummy4 := GetPluginByName("Dummy_4")

	//   -> 2
	// 1      -> 4
	//   -> 3
	graph := GetEmptyGraph()
	graph.AddNode(&dummy1)
	graph.AddNode(&dummy2)
	graph.AddNode(&dummy3)
	graph.AddNode(&dummy4)

	AddPath(&dummy1, []*Plugin{&dummy2, &dummy3})
	AddPath(&dummy2, []*Plugin{&dummy4})
	AddPath(&dummy3, []*Plugin{&dummy4})

	w := GetWorker()
	w.SetGraph(graph)

	w.RunGraph()

	unit := dummy4.fetchUnit()
	cnt, _ := unit.GetBuf().(int)

	assert.Equal(t, 40002, cnt, "Count should be 40002")
}

func TestGraphMultipleOutput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := GetPluginByName("Dummy_root")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")
	dummy4 := GetPluginByName("Dummy_4")

	//        -> 3
	// 1 -> 2
	//        -> 4
	graph := GetEmptyGraph()
	graph.AddNode(&dummy1)
	graph.AddNode(&dummy2)
	graph.AddNode(&dummy3)
	graph.AddNode(&dummy4)

	AddPath(&dummy1, []*Plugin{&dummy2})
	AddPath(&dummy2, []*Plugin{&dummy3, &dummy4})

	w := GetWorker()
	w.SetGraph(graph)

	w.RunGraph()

	unit1 := dummy3.fetchUnit()
	cnt1, _ := unit1.GetBuf().(int)

	assert.Equal(t, 20001, cnt1, "First count should be 20001")

	unit2 := dummy4.fetchUnit()
	cnt2, _ := unit2.GetBuf().(int)

	assert.Equal(t, 20001, cnt2, "Second count should be 20001")
}

func TestGraphBuilding(t *testing.T) {
	// Since graph uses pointers to store plugins, we cannot compare the constructed graph with one built manually
	// As an alternative, we use representative fields to compare the graph
	dummyParam1 := ConstructOverallParam("Dummy_1", "{}", []string{"Dummy_2"})
	dummyParam2 := ConstructOverallParam("Dummy_2", "{}", []string{"Dummy_3"})
	dummyParam3 := ConstructOverallParam("Dummy_3", "{}", []string{})

	builtOutput := buildGraph([]OverallParams{dummyParam1, dummyParam2, dummyParam3})

	// Check each node
	for idx, pg := range builtOutput.nodes {
		pgName := "Dummy_" + strconv.Itoa(idx+1)
		if pg.Name != pgName {
			panic(fmt.Sprintf("Name not match. Expecting %s, but got %s", pgName, pg.Name))
		}
		assert.Equal(t, "{}", pg.m_parameter)
		if idx != 2 {
			if len(pg.children) != 1 {
				panic(fmt.Sprintf("Output edge count not match. Expecting 1, but got %d", len(pg.children)))
			}
			nextName := "Dummy_" + strconv.Itoa(idx+2)
			if pg.children[0].Name != nextName {
				panic(fmt.Sprintf("Children edge not match. Expecting %s, but got %s", pg.children[0].Name, nextName))
			}
		}
	}
}
