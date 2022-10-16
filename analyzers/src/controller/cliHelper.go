package controller

type command struct {
	option    string // Possible set of options mapping to this command
	variable  string // Variables corresponding to the command
	action    func()   // What to do
}

type CliParser struct {
	commands  []command
}

func (parser *CliParser) AddCommand(option string, variable string, action func()) {
	parser.commands = append(parser.commands, command{option: option, variable: variable, action: action})
}

func (parser *CliParser) Parse(args []string) {

}

func GetParser() CliParser {
	return CliParser{commands: make([]command, 0)}
}
