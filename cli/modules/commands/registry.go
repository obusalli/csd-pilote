package commands

// Command represents a CLI command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Handler     func(args []string)
	SubCommands []SubCommand
}

// SubCommand represents a sub-command
type SubCommand struct {
	Name        string
	Description string
}

var commands = make(map[string]*Command)

// RegisterCommand registers a command
func RegisterCommand(cmd *Command) {
	commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		commands[alias] = cmd
	}
}

// GetCommand returns a command by name
func GetCommand(name string) *Command {
	return commands[name]
}

// GetAllCommands returns all registered commands
func GetAllCommands() map[string]*Command {
	return commands
}

func init() {
	// Register all commands
	RegisterCommand(DatabaseCommand())
	RegisterCommand(SeedCommand())
}
