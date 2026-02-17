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
	flags       []*Flag
	minArg      int
	maxArg      int
	action      func(args []string, flags Flags)
	parent      *Command
}

// AddCommand creates a new top-level command in the application.
func (a *App) AddCommand(name string) *Command {
	c := &Command{
		name:        name,
		flags:       []*Flag{},
		subcommands: []*Command{},
	}
	a.commands = append(a.commands, c)
	return c
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

// WithMaxArg sets the maximum number of positional arguments allowed for the command.
func (c *Command) WithMaxArg(max int) *Command {
	c.maxArg = max
	return c
}

// AddSubcommand creates a new command under the current command.
// It returns the newly created subcommand for chaining.
func (c *Command) AddSubcommand(name string) *Command {
	sub := &Command{name: name, parent: c}
	c.subcommands = append(c.subcommands, sub)
	return sub
}

// Ok finalizes the command, and returns to the parent command.
func (c *Command) Ok() *Command {
	return c.parent
}

// Action assigns the function to be executed when the command is run.
// It receives positional arguments and the parsed flags.
func (c *Command) Action(fn func(args []string, flags Flags)) *Command {
	c.action = fn
	return c
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

// Flags returns the list of registered flags.
func (c *Command) Flags() []*Flag { return c.flags }

// MinArg returns the minimum number of required positional arguments.
func (c *Command) MinArg() int { return c.minArg }

// MaxArg returns the maximum number of allowed positional arguments.
func (c *Command) MaxArg() int { return c.maxArg }

// Parent returns the parent command of the command.
func (c *Command) Parent() *Command { return c.parent }
