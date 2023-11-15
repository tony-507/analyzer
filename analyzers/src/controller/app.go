package controller

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/tony-507/analyzers/src/common/logging"
	"github.com/tony-507/analyzers/src/tttKernel"
)

type _VARTYPE int

const (
	_VAR_PLUGIN _VARTYPE = 1
	_VAR_VALUE  _VARTYPE = 2
)

type scriptVar struct {
	name       string
	varType    _VARTYPE
	value      string // Array value not supported
	attributes []*scriptVar
}

func (v *scriptVar) getAttributeStr() string {
	s := ""
	if len(v.attributes) != 0 || v.varType == _VAR_PLUGIN {
		s = "{"
		fieldArr := make([]string, 0)
		for _, field := range v.attributes {
			fieldArr = append(fieldArr, "\""+field.name+"\":"+field.getAttributeStr())
		}
		s += strings.Join(fieldArr, ",") + "}"
	} else {
		_, err := strconv.Atoi(v.value)
		if err != nil && v.value != "true" && v.value != "false" {
			s = "\"" + v.value + "\""
		} else {
			s = v.value
		}
	}
	return s
}

// Migration in progress
type tttController struct {
	logger      logging.Log
	aliasMap    map[string]string
	edgeMap     map[string][]string
	variables   []*scriptVar
	description string
}

var AppVersion = "unknown"

func Version() string {
	return AppVersion
}

func ListApp(resourceDir string) {
	fileInfo, err := ioutil.ReadDir(resourceDir)
	if err != nil {
		panic(err)
	}
	for _, file := range fileInfo {
		ctrl := getKernel()
		appName := strings.Split(file.Name(), ".")[0]
		ctrl.buildParams(getApp(resourceDir, appName), []string{}, 1)
		fmt.Println(fmt.Sprintf("%10s%10s%50s", appName, " ", ctrl.description))
	}
}

func StartApp(resourceDir string, appName string, input []string) {
	ctrl := getKernel()

	ctrl.buildParams(getApp(resourceDir, appName), input, -1)

	provider := tttKernel.NewWorker()

	pluginParams := make([]tttKernel.OverallParams, 0)
	for _, v := range ctrl.variables {
		if v.varType == _VAR_PLUGIN {
			pluginParams = append(pluginParams, tttKernel.ConstructOverallParam(v.value, v.getAttributeStr(), ctrl.edgeMap[v.name]))
		}
	}

	provider.StartService(pluginParams, selectPlugin)
}

func getKernel() tttController {
	return tttController{
		logger:    logging.CreateLogger("Controller"),
		aliasMap:  map[string]string{},
		edgeMap:   map[string][]string{},
		variables: []*scriptVar{},
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

// Read from script and input to prepare plugins and the respective parameters
func (ctrl *tttController) buildParams(script string, input []string, lim int) {
	if lim < 0 {
		lim = 99999
	}
	lines := strings.FieldsFunc(script, func(r rune) bool { return r == ';' || r == '\n' })
	syntaxErrLine := 0
	msg := ""
	conditionalStack := make([]bool, 0)
	for lNum, line := range lines {
		if lNum >= lim {
			break
		}
		line = strings.TrimSpace(line)
		// Skip empty line
		if len(line) == 0 {
			continue
		}

		// Handle functions, declarations and conditionals
		origTokens := strings.FieldsFunc(line, func(r rune) bool { return r == '(' || r == ',' || r == ')' || r == '|' || r == '=' || r == ' ' })

		tokens := make([]string, 0)
		// Remove spaces
		for idx := range origTokens {
			origTokens[idx] = strings.TrimSpace(origTokens[idx])
			if len(origTokens[idx]) > 0 {
				tokens = append(tokens, origTokens[idx])
			}
		}

		runLine := true
		if len(conditionalStack) > 0 {
			runLine = conditionalStack[len(conditionalStack)-1]
		}

		switch tokens[0] {
		// Built-in functions
		case "link":
			if runLine {
				if len(tokens) != 3 {
					panic(fmt.Sprintf("Wrong number of input: %d, %v", len(tokens), tokens))
				}
				if _, hasKey := ctrl.edgeMap[tokens[1]]; hasKey {
					ctrl.edgeMap[tokens[1]] = append(ctrl.edgeMap[tokens[1]], ctrl.getValueFromName(tokens[2]))
				} else {
					ctrl.edgeMap[tokens[1]] = []string{ctrl.getValueFromName(tokens[2])}
				}
			}
		case "alias":
			if runLine {
				ctrl.handleAliasing(tokens[1], tokens[2])
			}
		// Control flows
		case "if":
			if runLine {
				initFrom := []rune(tokens[1])
				switch initFrom[0] {
				case '$':
					if len(tokens) != 3 {
						tokens = append(tokens, "")
					}
					runConditional := ctrl.getValueFromArgs(input, string(initFrom)[1:], tokens[2]) != ""
					conditionalStack = append(conditionalStack, runConditional)
				default:
					panic("Unknown condition for if block")
				}
			}
		case "end":
			if len(conditionalStack) > 1 {
				conditionalStack = conditionalStack[:len(conditionalStack)-1]
			} else {
				conditionalStack = make([]bool, 0)
			}
		case "//":
			ctrl.description = strings.Join(tokens[1:], " ")
		// Declarations
		default:
			if runLine {
				newVar := scriptVar{}
				initFrom := []rune(tokens[1])
				if len(tokens) != 3 {
					tokens = append(tokens, "")
				}
				switch initFrom[0] {
				// #: declare plugin
				case '#':
					newVar.name = tokens[0]
					newVar.value = string(initFrom[1:])
					newVar.varType = _VAR_PLUGIN
					newVar.attributes = make([]*scriptVar, 0)
					ctrl.variables = append(ctrl.variables, &newVar)
				// $: User input
				case '$':
					splitName := strings.Split(tokens[0], ".")
					existVar := ctrl.getVariable(splitName, false)
					if existVar != nil {
						existVar.value = ctrl.getValueFromArgs(input, string(initFrom[1:]), tokens[2])
					} else {
						newVar.varType = _VAR_VALUE
						newVar.value = ctrl.getValueFromArgs(input, string(initFrom[1:]), tokens[2])
						newVar.attributes = make([]*scriptVar, 0)
						if len(splitName) > 1 {
							newVar.name = splitName[len(splitName)-1]
							parentVar := ctrl.getVariable(splitName[:len(splitName)-1], true)
							parentVar.attributes = append(parentVar.attributes, &newVar)
						} else {
							newVar.name = tokens[0]
							ctrl.variables = append(ctrl.variables, &newVar)
						}
					}
				default:
					splitName := strings.Split(tokens[0], ".")
					existVar := ctrl.getVariable(splitName, false)
					if existVar != nil {
						existVar.value = ctrl.resolveRHS(tokens[1])
					} else {
						newVar.varType = _VAR_VALUE
						newVar.value = ctrl.resolveRHS(tokens[1])
						newVar.attributes = make([]*scriptVar, 0)
						if len(splitName) > 1 {
							newVar.name = splitName[len(splitName)-1]
							parentVar := ctrl.getVariable(splitName[:len(splitName)-1], true)
							parentVar.attributes = append(parentVar.attributes, &newVar)
						} else {
							newVar.name = tokens[0]
							ctrl.variables = append(ctrl.variables, &newVar)
						}
					}
				}
			}
		}
	}

	if syntaxErrLine > 0 {
		panic(fmt.Sprintf("Line %d: %s", syntaxErrLine+1, msg))
	}
}

func (ctrl *tttController) handleAliasing(alias string, orig string) {
	ctrl.aliasMap[alias] = orig
}

func (ctrl *tttController) getVariable(components []string, createIfNeeded bool) *scriptVar {
	depth := len(components)
	curArrPtr := &ctrl.variables
	var match *scriptVar
	for i := 0; i < depth; i++ {
		arr := curArrPtr
		match = nil
		hasMatch := false
		for _, child := range *arr {
			if child.name == components[i] {
				hasMatch = true
				curArrPtr = &child.attributes
				match = child
			}
		}
		// If no match, create
		if !hasMatch && createIfNeeded {
			newVar := scriptVar{name: components[i], varType: _VAR_VALUE, attributes: make([]*scriptVar, 0)}
			*curArrPtr = append(*curArrPtr, &newVar)
			curArrPtr = &newVar.attributes
			match = &newVar
		} else if !hasMatch && !createIfNeeded {
			break
		}
	}
	return match
}

func (ctrl *tttController) getValueFromArgs(input []string, opt string, def string) string {
	for idx, param := range input {
		firstChar := []rune(param)[0]
		if firstChar == '-' {
			inOpt := strings.Trim(param, "-")
			if _, hasKey := ctrl.aliasMap[inOpt]; hasKey {
				inOpt = ctrl.aliasMap[inOpt]
			}
			if inOpt == opt && idx+1 < len(input) {
				return input[idx+1]
			}
		}
	}
	return def
}

func (ctrl *tttController) getValueFromName(name string) string {
	for _, v := range ctrl.variables {
		if v.name == name {
			return v.value
		}
	}
	return ""
}

// Determine if RHS is another variable
// If yes, return value of the variable (no recursion)
// else, return the string itself
func (ctrl *tttController) resolveRHS(rhs string) string {
	if x := ctrl.getValueFromName(rhs); x != "" {
		return x
	}
	return rhs
}
