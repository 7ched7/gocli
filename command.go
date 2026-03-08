package gocli

// Command represents a single CLI command.
// It includes command name, alias, descriptions, subcommands,
// flags, argument constraints, and an action function
// called when the command is run.
type Command struct {
	name        string
	alias       string
	short       string
	long        string
	subcommands []*Command
	flags       []FlagInfo
	minArg      int
	maxArg      int
	action      func(ctx *Context)
	parent      *Command
}

// NewCommand creates a new command with the given name.
func NewCommand(name string) *Command {
	return &Command{
		name:        name,
		flags:       []FlagInfo{},
		subcommands: []*Command{},
	}
}

// AddCommand registers a top-level command to the application.
func (a *App) AddCommand(command *Command) *App {
	command.parent = a.root
	a.commands = append(a.commands, command)
	return a
}

// WithAlias sets the short alias for the command.
// Alias provides alternative way for users to run the command.
func (c *Command) WithAlias(alias string) *Command {
	c.alias = alias
	return c
}

// WithShort sets the short description for the command.
// This is shown in commands section within help menu.
func (c *Command) WithShort(short string) *Command {
	c.short = short
	return c
}

// WithLong sets the detailed description for the command.
// This is shown in command help menu.
func (c *Command) WithLong(long string) *Command {
	c.long = long
	return c
}

// WithMinArg sets the minimum number of positional arguments required by the command.
func (c *Command) WithMinArg(min int) *Command {
	c.minArg = min
	return c
}

// WithMinArg sets the minimum number of positional arguments required by the application.
func (a *App) WithMinArg(min int) *App {
	a.root.minArg = min
	return a
}

// WithMaxArg sets the maximum number of positional arguments allowed for the command.
func (c *Command) WithMaxArg(max int) *Command {
	c.maxArg = max
	return c
}

// WithMaxArg sets the maximum number of positional arguments allowed for the application.
func (a *App) WithMaxArg(max int) *App {
	a.root.maxArg = max
	return a
}

// AddSubcommand registers a subcommand to the current command.
func (c *Command) AddSubcommand(subcommand *Command) *Command {
	subcommand.parent = c
	c.subcommands = append(c.subcommands, subcommand)
	return c
}

// Action assigns the function to be executed when the command is run.
// The handler receives a Context containing positional arguments and parsed flags.
func (c *Command) Action(fn func(ctx *Context)) *Command {
	c.action = fn
	return c
}

// Action assigns the default action to be executed when the application is run
// without specifying any command.
func (a *App) Action(fn func(ctx *Context)) *App {
	a.root.action = fn
	return a
}

// Name returns the name of the command.
func (c *Command) Name() string { return c.name }

// Alias returns the alias of the command.
// If not set, it returns an empty string.
func (c *Command) Alias() string { return c.alias }

// Short returns the short description of the command.
// If not set, it returns an empty string.
func (c *Command) Short() string { return c.short }

// Long returns the long description of the command.
// If not set, it returns an empty string.
func (c *Command) Long() string { return c.long }

// Subcommands returns all subcommands registered under the command.
func (c *Command) Subcommands() []*Command { return c.subcommands }

// Flags returns the list of flags registered for the command.
func (c *Command) Flags() []FlagInfo { return c.flags }

// MinArg returns the minimum number of positional arguments required by the command.
// If not set, it returns 0.
func (c *Command) MinArg() int { return c.minArg }

// MinArg returns the minimum number of positional arguments required by the application.
// If not set, it returns 0.
func (a *App) MinArg() int { return a.root.minArg }

// MaxArg returns the maximum number of positional arguments allowed for the command.
// If not set, it returns 0.
func (c *Command) MaxArg() int { return c.maxArg }

// MaxArg returns the maximum number of positional arguments allowed for the application.
// If not set, it returns 0.
func (a *App) MaxArg() int { return a.root.maxArg }

// Parent returns the parent command in the hierarchy.
// If the command has no parent command, it returns nil.
func (c *Command) Parent() *Command {
	if c == nil {
		return nil
	}
	return c.parent
}
