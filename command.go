package gocli

// Command represents a single CLI command.
// It includes subcommands, flags, argument constraints,
// and an action function called when the command is run.
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
func (c *Command) WithAlias(alias string) *Command {
	c.alias = alias
	return c
}

// WithShort sets the short description for the command.
func (c *Command) WithShort(short string) *Command {
	c.short = short
	return c
}

// WithLong sets the detailed description for the command.
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
// It receives positional arguments and the parsed flags.
func (c *Command) Action(fn func(ctx *Context)) *Command {
	c.action = fn
	return c
}

// Action assigns the function to be executed when no command is entered.
// It receives positional arguments and the parsed flags.
func (a *App) Action(fn func(ctx *Context)) *App {
	a.root.action = fn
	return a
}

// Name returns the name of the command.
func (c *Command) Name() string { return c.name }

// Alias returns the alias of the command.
func (c *Command) Alias() string { return c.alias }

// Short returns the short description of the command.
func (c *Command) Short() string { return c.short }

// Long returns the long description of the command.
func (c *Command) Long() string { return c.long }

// Subcommands returns the list of nested subcommands.
func (c *Command) Subcommands() []*Command { return c.subcommands }

// Flags returns the list of registered flags of the command.
func (c *Command) Flags() []FlagInfo { return c.flags }

// MinArg returns the minimum number of required positional arguments of the command.
func (c *Command) MinArg() int { return c.minArg }

// MinArg returns the minimum number of required positional arguments of the application.
func (a *App) MinArg() int { return a.root.minArg }

// MaxArg returns the maximum number of allowed positional arguments for the command.
func (c *Command) MaxArg() int { return c.maxArg }

// MaxArg returns the maximum number of allowed positional arguments for the application.
func (a *App) MaxArg() int { return a.root.maxArg }

// Parent returns the parent command of the command.
func (c *Command) Parent() *Command { return c.parent }
