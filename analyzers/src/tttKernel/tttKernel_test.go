package tttKernel

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBufOverwriteField(t *testing.T) {
	buf := MakeSimpleBuf([]byte{})
	field := "dummy"

	buf.SetField(field, 0, true)

	v1, ok := buf.GetField(field)
	if !ok {
		panic("No such field")
	}
	assert.Equal(t, 0, v1, "Field should be 0")

	buf.SetField(field, 10, false)

	v2, ok := buf.GetField(field)
	if !ok {
		panic("No such field")
	}
	assert.Equal(t, 10, v2, "Field should be 10")
}

func dummySelector(inputName string) IPlugin {
	splitName := strings.Split(inputName, "_")
	role := 1
	if splitName[1] == "root" {
		role = 0
	}
	return dummy(inputName, role)
}

func TestPluginInterfaces(t *testing.T) {
	pg := getPluginByName("Dummy_1", dummySelector)
	pg.impl.SetCallback(func(string, WORKER_REQUEST, interface{}) {

	})

	inUnit := &dummyUnit{MakeSimpleBuf([]byte{byte(2)})}
	pg.impl.DeliverUnit(inUnit, "")

	unit := pg.impl.FetchUnit()

	rv := int(GetBytesInBuf(unit)[0])
	if rv != 20 {
		panic(fmt.Sprintf("Expect return value 20, but got %d", rv))
	}
}

func TestSimpleGraph(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root", dummySelector)
	dummy2 := getPluginByName("Dummy_2", dummySelector)
	dummy3 := getPluginByName("Dummy_3", dummySelector)

	nodes := []*graphNode{dummy1, dummy2, dummy3}

	// root -> 2 -> 3
	addPath(dummy1, []*graphNode{dummy2})
	addPath(dummy2, []*graphNode{dummy3})

	w := NewWorker()
	w.setGraph(nodes)

	w.runGraph()

	unit := dummy3.fetchUnit()
	cnt := int(GetBytesInBuf(unit)[0])

	assert.Equal(t, 33, cnt, "Count should be 33 (20001 casted to byte)")
}

func TestGraphMultipleInput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root", dummySelector)
	dummy2 := getPluginByName("Dummy_2", dummySelector)
	dummy3 := getPluginByName("Dummy_3", dummySelector)
	dummy4 := getPluginByName("Dummy_4", dummySelector)

	nodes := []*graphNode{dummy1, dummy2, dummy3, dummy4}
	//   -> 2
	// 1      -> 4
	//   -> 3

	addPath(dummy1, []*graphNode{dummy2, dummy3})
	addPath(dummy2, []*graphNode{dummy4})
	addPath(dummy3, []*graphNode{dummy4})

	w := NewWorker()
	w.setGraph(nodes)

	w.runGraph()

	unit := dummy4.fetchUnit()
	cnt := int(GetBytesInBuf(unit)[0])

	assert.Equal(t, 66, cnt, "Count should be 66 (40002 casted to byte)")
}

func TestGraphMultipleOutput(t *testing.T) {
	// Declare here to prevent dangling pointer
	dummy1 := getPluginByName("Dummy_root", dummySelector)
	dummy2 := getPluginByName("Dummy_2", dummySelector)
	dummy3 := getPluginByName("Dummy_3", dummySelector)
	dummy4 := getPluginByName("Dummy_4", dummySelector)

	nodeList := []*graphNode{dummy1, dummy2, dummy3, dummy4}
	//        -> 3
	// 1 -> 2
	//        -> 4
	addPath(dummy1, []*graphNode{dummy2})
	addPath(dummy2, []*graphNode{dummy3, dummy4})

	w := NewWorker()
	w.setGraph(nodeList)

	w.runGraph()

	unit1 := dummy3.fetchUnit()
	cnt1 := int(GetBytesInBuf(unit1)[0])

	assert.Equal(t, 33, cnt1, "First count should be 33 (20001 casted to byte)")

	unit2 := dummy4.impl.FetchUnit()
	cnt2 := int(GetBytesInBuf(unit2)[0])

	assert.Equal(t, 33, cnt2, "Second count should be 33 (20001 casted to byte)")
}

func TestGraphBuilding(t *testing.T) {
	// Since graph uses pointers to store plugins, we cannot compare the constructed graph with one built manually
	// As an alternative, we use representative fields to compare the graph
	dummyParam1 := ConstructOverallParam("Dummy_1", "{}", []string{"Dummy_2"})
	dummyParam2 := ConstructOverallParam("Dummy_2", "{}", []string{"Dummy_3"})
	dummyParam3 := ConstructOverallParam("Dummy_3", "{}", []string{})

	builtOutput := buildGraph([]OverallParams{dummyParam1, dummyParam2, dummyParam3}, dummySelector)

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
