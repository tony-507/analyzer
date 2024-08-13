package controller

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/tony-507/analyzers/src/logging"
	"github.com/tony-507/analyzers/src/tttKernel"
)

// Migration in progress
type tttController struct {
	logger logging.Log
	parser scriptParser
}

var AppVersion = "unknown"

func Version() string {
	return AppVersion
}

func LinkPlugin(parent *PluginBuilder, child *PluginBuilder) {
	parent.AddChild(child.name)
}

func LinkPlugins(plugins []*PluginBuilder) {
	for i := 0; i < len(plugins)-1; i++ {
		LinkPlugin(plugins[i], plugins[i+1])
	}
}

func Start(pluginParams *[]tttKernel.OverallParams, env *tttKernel.Resource) {
	provider := tttKernel.NewWorker()

	provider.UpdateResource(*env)

	provider.StartService(*pluginParams, selectPlugin)
}

func ListApp(resourceDir string) {
	fileInfo, err := ioutil.ReadDir(resourceDir)
	if err != nil {
		panic(err)
	}
	for _, file := range fileInfo {
		ctrl := newController()
		appName := strings.Split(file.Name(), ".")[0]
		ctrl.parser.buildParams(getApp(resourceDir, appName), []string{}, 1)
		fmt.Println(fmt.Sprintf("%10s%10s%50s", appName, " ", ctrl.parser.description))
	}
}

func StartApp(resourceDir string, appName string, input []string) {
	ctrl := newController()

	ctrl.parser.buildParams(getApp(resourceDir, appName), input, -1)

	provider := tttKernel.NewWorker()

	provider.UpdateResource(ctrl.parser.env)

	pluginParams := make([]tttKernel.OverallParams, 0)
	for _, v := range ctrl.parser.variables {
		if v.varType == _VAR_PLUGIN {
			pluginParams = append(pluginParams, tttKernel.ConstructOverallParam(v.value, v.getAttributeStr(), ctrl.parser.edgeMap[v.name]))
		}
	}

	provider.StartService(pluginParams, selectPlugin)
}

func newController() tttController {
	return tttController{
		logger: logging.CreateLogger("Controller"),
		parser: newScriptParser(),
	}
}

func getApp(resourceDir string, appName string) string {
	fileInfo, err := ioutil.ReadDir(resourceDir)
	if err != nil {
		panic(err)
	}
	rv := ""

	for _, file := range fileInfo {
		app := strings.Split(file.Name(), ".")[0]
		if app == appName {
			buf, err := ioutil.ReadFile(resourceDir + file.Name())
			if err != nil {
				panic(err)
			}
			rv = string(buf)
		}
	}

	return rv
}
