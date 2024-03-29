package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tony-507/analyzers/src/tttKernel"
)

// A ttt struct

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

// Script parsing

type scriptParser struct {
	description string              // Description
	aliasMap    map[string]string   // Aliases
	edgeMap     map[string][]string // Graph topology
	variables   []*scriptVar        // Variables
	env         tttKernel.Resource
}

func newScriptParser() scriptParser {
	return scriptParser{
		aliasMap: map[string]string{},
		edgeMap: map[string][]string{},
		variables: []*scriptVar{},
		env: tttKernel.Resource{
			OutDir: "output",
		},
	}
}

// Read from script and input to prepare plugins and the respective parameters
func (sp *scriptParser) buildParams(script string, input []string, lim int) {
	if lim < 0 {
		lim = 99999
	}
	lines := strings.FieldsFunc(script, func(r rune) bool { return r == ';' || r == '\n' })
	syntaxErrLine := 0
	msg := ""
	conditionalStack := make([]bool, 0) // Stack of if-blocks
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

		// Skip lines in unrelated if-block
		runLine := true
		if len(conditionalStack) > 0 {
			runLine = conditionalStack[len(conditionalStack)-1]
		}

		if tokens[0] == "end" {
			if len(conditionalStack) > 1 {
				conditionalStack = conditionalStack[:len(conditionalStack)-1]
			} else {
				conditionalStack = make([]bool, 0)
			}
			continue
		}

		if !runLine {
			continue
		}

		switch tokens[0] {
		// Built-in functions
		case "link":
			if len(tokens) != 3 {
				panic(fmt.Sprintf("Wrong number of input: %d, %v", len(tokens), tokens))
			}
			sp.linkPlugins(tokens[1], tokens[2])
		case "alias":
			sp.setAlias(tokens[1], tokens[2])
		// Control flows
		case "if":
			initFrom := []rune(tokens[1])
			switch initFrom[0] {
			case '$':
				if len(tokens) != 3 {
					tokens = append(tokens, "")
				}
				runConditional := sp.getValueFromArgs(input, string(initFrom[1:]), tokens[2]) != ""
				conditionalStack = append(conditionalStack, runConditional)
			default:
				panic("Unknown condition for if block")
			}
		case "//":
			sp.description = strings.Join(tokens[1:], " ")
		// Declarations
		default:
			if len(tokens) != 3 {
				tokens = append(tokens, "")
			}
			initFrom := []rune(tokens[1])
			switch initFrom[0] {
			case '#':
				sp.declarePlugin(tokens[0], string(initFrom[1:]))
			case '$':
				// User input
				sp.declareVariable(tokens[0], sp.getValueFromArgs(input, string(initFrom[1:]), tokens[2]))
			default:
				sp.declareVariable(tokens[0], sp.resolveRHS(tokens[1]))
			}
		}
	}

	if syntaxErrLine > 0 {
		panic(fmt.Sprintf("Line %d: %s", syntaxErrLine+1, msg))
	}
}

func (sp *scriptParser) setAlias(alias string, orig string) {
	sp.aliasMap[alias] = orig
}

func (sp *scriptParser) linkPlugins(parent string, child string) {
	if _, hasKey := sp.edgeMap[parent]; hasKey {
		sp.edgeMap[parent] = append(sp.edgeMap[parent], sp.getValueFromName(child))
	} else {
		sp.edgeMap[parent] = []string{sp.getValueFromName(child)}
	}
}

func (sp *scriptParser) declarePlugin(name string, pluginType string) {
	newVar := scriptVar{
		name: name,
		value: pluginType,
		varType: _VAR_PLUGIN,
		attributes: make([]*scriptVar, 0),
	}
	sp.variables = append(sp.variables, &newVar)
}

func (sp *scriptParser) declareVariable(fullname string, value string) {
	splitName := strings.Split(fullname, ".")
	existVar := sp.getVariable(splitName, false)
	if existVar != nil {
		existVar.value = value
	} else {
		if len(splitName) > 1 {
			if splitName[0] == "global" {
				sp.setGlobalParam(splitName[1], value)
			} else {
				parentVar := sp.getVariable(splitName[:len(splitName)-1], true)
				v := sp.newVariable(splitName[len(splitName)-1], value)
				parentVar.attributes = append(parentVar.attributes, &v)
			}
		} else {
			v := sp.newVariable(splitName[len(splitName)-1], value)
			sp.variables = append(sp.variables, &v)
		}
	}
}

func (sp *scriptParser) newVariable(name string, value string) scriptVar {
	return scriptVar{
		name: name,
		value: value,
		varType: _VAR_VALUE,
		attributes: make([]*scriptVar, 0),
	}
}

func (sp *scriptParser) setGlobalParam(name string, value string) {
	switch (name) {
	case "outDir":
		sp.env.OutDir = value
	}
}

func (sp *scriptParser) getVariable(components []string, createIfNeeded bool) *scriptVar {
	depth := len(components)
	curArrPtr := &sp.variables
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

func (sp *scriptParser) getValueFromArgs(input []string, opt string, def string) string {
	for idx, param := range input {
		firstChar := []rune(param)[0]
		if firstChar == '-' {
			inOpt := strings.Trim(param, "-")
			if _, hasKey := sp.aliasMap[inOpt]; hasKey {
				inOpt = sp.aliasMap[inOpt]
			}
			if inOpt == opt && idx+1 < len(input) {
				return input[idx+1]
			}
		}
	}
	return def
}

func (sp *scriptParser) getValueFromName(name string) string {
	for _, v := range sp.variables {
		if v.name == name {
			return v.value
		}
	}
	return ""
}

// Determine if RHS is another variable
// If yes, return value of the variable (no recursion)
// else, return the string itself
func (sp *scriptParser) resolveRHS(rhs string) string {
	if x := sp.getValueFromName(rhs); x != "" {
		return x
	}
	return rhs
}
