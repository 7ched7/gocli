package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, expected value type, default value,
// and description for help menu.
type Flag struct {
	name         string   // Flag name
	alias        string   // Optional flag alias
	flagType     flagType // Flag value type
	defaultValue any      // Default value and type
	description  string   // Description shown in help
}

// flagBuilder is a fluent builder used for assigning flags to a command.
type flagBuilder struct {
	target *Flag
	parent *Command
}

// AddFlag creates a new flag for a command.
// It returns flagBuilder that provides a fluent interface.
func (c *Command) AddFlag(name string) *flagBuilder {
	o := &Flag{name: name}
	return &flagBuilder{
		target: o,
		parent: c,
	}
}

// WithAlias sets the alias for the flag.
func (o *flagBuilder) WithAlias(alias string) *flagBuilder {
	o.target.alias = alias
	return o
}

// WithType sets the type for the flag.
func (o *flagBuilder) WithType(flagType flagType) *flagBuilder {
	o.target.flagType = flagType
	return o
}

// WithDefault sets the default value for the flag.
func (o *flagBuilder) WithDefault(defaultValue any) *flagBuilder {
	o.target.defaultValue = defaultValue
	return o
}

// WithDescription sets the description for the flag.
func (o *flagBuilder) WithDescription(description string) *flagBuilder {
	o.target.description = description
	return o
}

// Ok finalizes the flag, adds it to the parent command,
// and returns the parent command for chaining.
func (o *flagBuilder) Ok() *Command {
	o.parent.flags = append(o.parent.flags, o.target)
	return o.parent
}

// Name returns the name of the flag.
func (o *Flag) Name() string { return o.name }

// Alias returns the alias of the flag.
func (o *Flag) Alias() string { return o.alias }

// Type returns the type of the flag.
func (o *Flag) Type() flagType { return o.flagType }

// Default returns the default value of the flag.
func (o *Flag) Default() any { return o.defaultValue }

// Description returns the description of the flag.
func (o *Flag) Description() string { return o.description }
