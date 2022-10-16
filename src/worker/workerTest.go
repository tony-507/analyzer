package worker

import (
	"errors"
	"fmt"
	"strconv"

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
		err = testUtils.Assert_type_equal(pg.Work, "*worker.DummyPlugin")
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

func pluginInterfaceTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("pluginInterfaceTest")

	// Declare here to prevent dangling pointer
	dummy1 := GetPluginByName("Dummy_1")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")
	dummy4 := GetPluginByName("Dummy_4")
	dummy5 := GetPluginByName("Dummy_5")
	dummy6 := GetPluginByName("Dummy_6")

	// Construct graph now
	graph := GetEmptyGraph()
	graph.AddNode(&dummy1)
	graph.AddNode(&dummy2)
	graph.AddNode(&dummy3)
	graph.AddNode(&dummy4)
	graph.AddNode(&dummy5)
	graph.AddNode(&dummy6)

	AddPath(&dummy1, []*Plugin{&dummy2, &dummy3, &dummy4})
	AddPath(&dummy2, []*Plugin{&dummy5})
	AddPath(&dummy3, []*Plugin{&dummy5})
	AddPath(&dummy4, []*Plugin{&dummy6})
	AddPath(&dummy5, []*Plugin{&dummy6})

	w := GetWorker()
	w.SetGraph(graph)

	tc.Describe("Deliver to root", func(input interface{}) (interface{}, error) {
		unit := common.IOUnit{Buf: 20, IoType: 0, Id: 0}
		_, err := dummy1.DeliverUnit(unit)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	tc.Describe("Check fetch count of root", func(input interface{}) (interface{}, error) {
		unit := dummy1.FetchUnit()

		cnt, _ := unit.GetBuf().(int)
		cnt = cnt % 10
		err := testUtils.Assert_equal(cnt, 1)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	tc.Describe("Check fetch count of leaf", func(input interface{}) (interface{}, error) {
		unit := dummy6.FetchUnit()
		cnt, _ := unit.GetBuf().(int)
		err := testUtils.Assert_equal(cnt, 620013)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return tc
}

func workerRunGraphTest() testUtils.Testcase {
	tc := testUtils.GetTestCase("workerRunGraphTest")

	// Declare here to prevent dangling pointer
	dummy1 := GetPluginByName("Dummy_root")
	dummy2 := GetPluginByName("Dummy_2")
	dummy3 := GetPluginByName("Dummy_3")
	dummy4 := GetPluginByName("Dummy_4")

	tc.Describe("Graph with one input edge and one output edge", func (input interface{}) (interface{}, error) {
		// Construct graph now
		graph := GetEmptyGraph()
		graph.AddNode(&dummy1)
		graph.AddNode(&dummy2)
		graph.AddNode(&dummy3)

		AddPath(&dummy1, []*Plugin{&dummy2})
		AddPath(&dummy2, []*Plugin{&dummy3})

		dummy1.setParameterStr(0)
		dummy2.setParameterStr(1)
		dummy3.setParameterStr(1)

		w := GetWorker()
		w.SetGraph(graph)

		w.RunGraph()

		unit := dummy3.FetchUnit()
		cnt, _ := unit.GetBuf().(int)

		err := testUtils.Assert_equal(cnt, 20001)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	tc.Describe("Graph with multiple input edges", func (input interface{}) (interface{}, error) {
		// Construct graph now
		graph := GetEmptyGraph()
		graph.AddNode(&dummy1)
		graph.AddNode(&dummy2)
		graph.AddNode(&dummy3)
		graph.AddNode(&dummy4)

		AddPath(&dummy1, []*Plugin{&dummy2, &dummy3})
		AddPath(&dummy2, []*Plugin{&dummy4})
		AddPath(&dummy3, []*Plugin{&dummy4})

		dummy1.setParameterStr(0)
		dummy2.setParameterStr(1)
		dummy3.setParameterStr(1)
		dummy4.setParameterStr(2)

		w := GetWorker()
		w.SetGraph(graph)

		w.RunGraph()

		unit := dummy4.FetchUnit()
		cnt, _ := unit.GetBuf().(int)

		err := testUtils.Assert_equal(cnt, 4608225)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	tc.Describe("Graph with multiple output edges", func (input interface{}) (interface{}, error) {
		// Construct graph now
		graph := GetEmptyGraph()
		graph.AddNode(&dummy1)
		graph.AddNode(&dummy2)
		graph.AddNode(&dummy3)
		graph.AddNode(&dummy4)

		AddPath(&dummy1, []*Plugin{&dummy2})
		AddPath(&dummy2, []*Plugin{&dummy3, &dummy4})

		dummy1.setParameterStr(0)
		dummy2.setParameterStr(1)
		dummy3.setParameterStr(1)
		dummy4.setParameterStr(1)

		w := GetWorker()
		w.SetGraph(graph)

		w.RunGraph()

		unit1 := dummy3.FetchUnit()
		cnt1, _ := unit1.GetBuf().(int)

		err := testUtils.Assert_equal(cnt1, 1514212)
		if err != nil {
			return nil, err
		}

		unit2 := dummy4.FetchUnit()
		cnt2, _ := unit2.GetBuf().(int)

		err = testUtils.Assert_equal(cnt2, 68156039)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return tc
}

func graphBuildingTest() testUtils.Testcase  {
	// Since graph uses pointers to store plugins, we cannot compare the constructed graph with one built manually
	// As an alternative, we use representative fields to compare the graph
	tc := testUtils.GetTestCase("GraphBuildingTest")

	tc.Describe("Build graph in which each node has one input edge and one output edge", func (input interface{}) (interface{}, error) {
		dummyParam1 := ConstructOverallParam("Dummy_1", 1, []string{"Dummy_2"})
		dummyParam2 := ConstructOverallParam("Dummy_2", 2, []string{"Dummy_3"})
		dummyParam3 := ConstructOverallParam("Dummy_3", 3, []string{})

		builtOutput := buildGraph([]OverallParams{dummyParam1, dummyParam2, dummyParam3})

		// Check each node
		for idx, pg := range builtOutput.nodes {
			pgName := "Dummy_" + strconv.Itoa(idx + 1)
			if (pg.Name != pgName) {
				msg := fmt.Sprintf("Name not match. Expecting %s, but got %s", pgName, pg.Name)
				return nil, errors.New(msg)
			}
			param, isInt := pg.m_parameter.(int)
			if !isInt {
				return nil, errors.New("Parameter is not an integer")
			}
			if param != idx + 1 {
				msg := fmt.Sprintf("Parameter not match. Expecting %s, but got %s", strconv.Itoa(idx + 1),  strconv.Itoa(param))
				return nil, errors.New(msg)
			}
			if idx != 2 {
				if len(pg.children) != 1 {
					msg := fmt.Sprintf("Output edge count not match. Expecting 1, but got %s", strconv.Itoa(len(pg.children)))
					return nil, errors.New(msg)
				}
				nextName := "Dummy_" + strconv.Itoa(idx + 2)
				if pg.children[0].Name != nextName {
					msg := fmt.Sprintf("Children edge not match. Expecting %s, but got %s", pg.children[0].Name, nextName)
					return nil, errors.New(msg)
				}
			}
		}
		return nil, nil
	})

	return tc
}

func AddUnitTestSuite(t *testUtils.Tester) {
	tmg := testUtils.GetTestCaseMgr()

	// We may add custom test filter here later

	tmg.AddTest(pluginUnitTest, []string{"worker"})
	tmg.AddTest(pluginInterfaceTest, []string{"worker"})
	tmg.AddTest(workerRunGraphTest, []string{"worker"})
	tmg.AddTest(graphBuildingTest, []string{"worker"})

	t.AddSuite("unitTest", tmg)
}
