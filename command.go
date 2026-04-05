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
	subcommands []CommandInfo
	flags       []FlagInfo
	minArg      int
	maxArg      int
	actionF     func(ctx *Context)
	parent      CommandInfo
}

// CommandInfo provides access to command metadata and behaviour.
type CommandInfo interface {
	Name() string               // Name returns the name of the command.
	Alias() string              // Alias returns the alias of the command.
	Short() string              // Short returns the short description of the command.
	Long() string               // Long returns the long description of the command.
	Subcommands() []CommandInfo // Subcommands returns all subcommands registered under the command.
	Flags() []FlagInfo          // Flags returns the list of flags registered for the command.
	MinArg() int                // MinArg returns the minimum number of positional arguments.
	MaxArg() int                // MaxArg returns the maximum number of positional arguments.
	Parent() CommandInfo        // Parent returns the parent command in the hierarchy.

	action() func(ctx *Context)
}

// NewCommand creates a new command with the given name.
func NewCommand(name string) *Command {
	return &Command{
		name:        name,
		flags:       []FlagInfo{},
		subcommands: []CommandInfo{},
	}
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

// WithMaxArg sets the maximum number of positional arguments allowed for the command.
func (c *Command) WithMaxArg(max int) *Command {
	c.maxArg = max
	return c
}

// Action assigns the function to be executed when the command is run.
// The handler receives a Context containing positional arguments and parsed flags.
func (c *Command) Action(fn func(ctx *Context)) *Command {
	c.actionF = fn
	return c
}

// AddFlag registers flags to the command.
func (c *Command) AddFlag(flags ...FlagInfo) *Command {
	for _, f := range flags {
		c.flags = append(c.flags, f)
	}
	return c
}

// AddSubcommand registers subcommands to the current command.
func (c *Command) AddSubcommand(commands ...*Command) *Command {
	for _, cmd := range commands {
		cmd.parent = c
		c.subcommands = append(c.subcommands, cmd)
	}
	return c
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
func (c *Command) Subcommands() []CommandInfo { return c.subcommands }

// Flags returns the list of flags registered for the command.
func (c *Command) Flags() []FlagInfo { return c.flags }

// MinArg returns the minimum number of positional arguments required by the command.
// If not set, it returns 0.
func (c *Command) MinArg() int { return c.minArg }

// MaxArg returns the maximum number of positional arguments allowed for the command.
// If not set, it returns 0.
func (c *Command) MaxArg() int { return c.maxArg }

// Parent returns the parent command in the hierarchy.
// If the command has no parent command, it returns nil.
func (c *Command) Parent() CommandInfo {
	if c == nil {
		return nil
	}
	return c.parent
}

func (c *Command) action() func(ctx *Context) { return c.actionF }
