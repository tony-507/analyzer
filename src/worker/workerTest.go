package worker

import (
	"errors"
	"fmt"

	"github.com/tony-507/analyzers/src/common"
	"github.com/tony-507/analyzers/test/testUtils"
)

func pluginUnitTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("pluginTest")

	tc.Describe("Initialization", func(input interface{}) (interface{}, error) {
		return "Dummy_1", nil
	})

	tc.Describe("Deduce plugin type", func(input interface{}) (interface{}, error) {
		cellName, isStr := input.(string)
		var err error
		if !isStr {
			err := errors.New("cellName is not a string")
			return nil, err
		}
		pg := GetPluginByName(cellName)
		_, err = testUtils.Assert_type_equal(pg.Work, "worker.DummyPlugin")
		if err != nil {
			return nil, err
		}
		return pg, nil
	})

	tc.Describe("Run DeliverUnit interface", func(input interface{}) (interface{}, error) {
		pg, isPlugin := input.(Plugin)
		if !isPlugin {
			err := errors.New("input is not a plugin")
			return nil, err
		}
		inUnit := common.IOUnit{Buf: 2, IoType: 0, Id: 0}
		_, err := pg.DeliverUnit(inUnit)
		if err != nil {
			return nil, err
		}
		return pg, nil
	})

	tc.Describe("Run FetchUnit interface", func(input interface{}) (interface{}, error) {
		pg, isPlugin := input.(Plugin)
		if !isPlugin {
			err := errors.New("input is not a plugin")
			return nil, err
		}
		unit := pg.FetchUnit()

		rv, isInt := unit.GetBuf().(int)
		if !isInt {
			err := errors.New("return value is not an integer")
			return nil, err
		}
		if rv != 20 {
			errMsg := fmt.Sprintf("Expect return value 20, but got %d", rv)
			err := errors.New(errMsg)
			return nil, err
		}
		return pg, nil
	})

	return tc
}

func workerRunGraphTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("WorkerRunGraphTest")

	// Declare here to prevent dangling pointer
	dummy1 := GetPluginByName("Dummy_1")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")
	dummy4 := GetPluginByName("Dummy_4")
	dummy5 := GetPluginByName("Dummy_5")
	dummy6 := GetPluginByName("Dummy_6")

	// Create nodes
	node1 := CreateNode(&dummy1, nil)
	node2 := CreateNode(&dummy2, nil)
	node3 := CreateNode(&dummy3, nil)
	node4 := CreateNode(&dummy4, nil)
	node5 := CreateNode(&dummy5, nil)
	node6 := CreateNode(&dummy6, nil)

	// Construct graph now
	graph := GetEmptyGraph()
	graph.AddRoot(&node1)

	AddPath(&node1, []*GraphNode{&node2, &node3, &node4})
	AddPath(&node2, []*GraphNode{&node5})
	AddPath(&node3, []*GraphNode{&node5})
	AddPath(&node4, []*GraphNode{&node6})
	AddPath(&node5, []*GraphNode{&node6})

	w := GetWorker()
	w.SetGraph(graph)

	tc.Describe("Deliver to root", func(input interface{}) (interface{}, error) {
		root := w.graph.GetRoots()[0]

		unit := common.IOUnit{Buf: 20, IoType: 0, Id: 0}
		_, err := root.GetVal().DeliverUnit(unit)
		if err != nil {
			return nil, err
		}
		return root, nil
	})

	tc.Describe("Check fetch count of root", func(input interface{}) (interface{}, error) {
		root, _ := input.(*GraphNode)
		unit := root.GetVal().FetchUnit()

		cnt, _ := unit.GetBuf().(int)
		cnt = cnt % 10
		_, err := testUtils.Assert_equal(cnt, 1)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	tc.Describe("Check fetch count of leaf", func(input interface{}) (interface{}, error) {
		unit := node6.val.FetchUnit()
		cnt, _ := unit.GetBuf().(int)
		_, err := testUtils.Assert_equal(cnt, 620013)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	tmg.AddTest(pluginUnitTest, []string{"worker"})
	tmg.AddTest(workerRunGraphTest, []string{"worker"})

	t.AddSuite("unitTest", tmg)
}
