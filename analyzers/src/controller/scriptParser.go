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
	lines := strings.FieldsFunc(script, func(r rune) bool { return r == ';' || r == '\n' })
	syntaxErrLine := 0
	msg := ""

	lNum := 0
	shouldRun := true

	// For loop handling
	loopVar := ""
	loopMax := -1
	loopIdx := -1

	if lim < 0 {
		lim = len(lines)
	}

	for lNum < len(lines) {
		if lNum >= lim {
			break
		}

		tokens := parseLine(lines[lNum])

		if len(tokens) == 0 || (!shouldRun && tokens[0] != "end") {
			lNum++
			continue
		}

		switch tokens[0] {
		case "//":
			// Description
			sp.description = strings.Join(tokens[1:], " ")
		case "link":
			// Link two plugins
			if len(tokens) != 3 {
				panic(fmt.Sprintf("Wrong number of input: %d, %v", len(tokens), tokens))
			}
			sp.linkPlugins(tokens[1], tokens[2])
		case "alias":
			// Set alias to a variable
			sp.setAlias(tokens[1], tokens[2])
		case "if":
			// Conditional
			initFrom := []rune(tokens[1])
			switch initFrom[0] {
			case '$':
				if len(tokens) != 3 {
					tokens = append(tokens, "")
				}
				shouldRun = sp.getValueFromArgs(input, string(initFrom[1:]), tokens[2], loopIdx) != ""
			default:
				syntaxErrLine = lNum
				msg = "Unknown condition for if block"
				break
			}
		case "for":
			// For loop
			loopVar = tokens[1]

			input = append(input, loopVar)
			input = append(input, strconv.Itoa(loopIdx))

			split := strings.Split(tokens[3], "..")
			loopIdx, _ = strconv.Atoi(split[0])
			loopMax, _ = strconv.Atoi(split[1])
		case "end":
			// End of control flow
			shouldRun = true
			if loopMax != -1 {
				loopIdx++

				input = input[:len(input)-2]

				if loopIdx == loopMax {
					loopVar = ""
					loopMax = -1
					loopIdx = -1
				} else {
					input = append(input, loopVar)
					input = append(input, strconv.Itoa(loopIdx))
				}
			}

		default:
			// Declaration
			sp.declare(tokens, input, loopIdx)
		}

		lNum++
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

/*
 * Declare a variable
 * - tokens: array in the form of [name, value, default]
 * - input: user input in the form of [flag, value, flag, value, ...]
 * - idx: index in for loop, default to be -1
 */
func (sp *scriptParser) declare(tokens []string, input []string, idx int) {
	if len(tokens) != 3 {
		tokens = append(tokens, "")
	}

	name, value, def := tokens[0], tokens[1], tokens[2]

	initFrom := []rune(value)

	target := string(initFrom[1:])

	switch initFrom[0] {
	case '#':
		sp.declarePlugin(name, target)
	case '$':
		// User input
		sp.declareVariable(name, sp.getValueFromArgs(input, target, def, idx))
	default:
		sp.declareVariable(name, sp.resolveRHS(tokens[1]))
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

func (sp *scriptParser) getValueFromArgs(input []string, opt string, def string, idx int) string {
	// Try "$opt"
	v_default := sp.getValueWithName(input, opt)
	if v_default != "" {
		return v_default
	}

	if idx >= 0 {
		// Try "$opt_$idx"
		v_idx := sp.getValueWithName(input, opt+"_"+strconv.Itoa(idx))
		if v_idx != "" {
			return v_idx
		}
	}

	return def
}

func (sp *scriptParser) getValueWithName(input []string, opt string) string {
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
	return ""
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

func parseLine(line string) []string {
	origTokens := strings.FieldsFunc(
		line,
		func(r rune) bool {
			return r == '(' || r == ',' || r == ')' || r == '|' || r == '=' || r == ' '
		},
	)

	tokens := make([]string, 0)
	for idx := range origTokens {
		origTokens[idx] = strings.TrimSpace(origTokens[idx])
		if len(origTokens[idx]) > 0 {
			tokens = append(tokens, origTokens[idx])
		}
	}

	return tokens
}
