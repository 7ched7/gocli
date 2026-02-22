package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, expected value type, default value,
// and description for help menu.
type Flag struct {
	name          string
	alias         string
	flagType      flagType
	defaultValue  any
	description   string
	allowedValues []string
}

// NewFlag creates a new flag with the given name.
func (c *App) NewFlag(name string) *Flag {
	return &Flag{name: name}
}

// AddFlag registers a flag to the command.
func (c *Command) AddFlag(flag *Flag) *Command {
	c.flags = append(c.flags, flag)
	return c
}

// AddGlobalFlag registers a global flag to the application.
func (a *App) AddGlobalFlag(flag *Flag) *App {
	a.globalFlags = append(a.globalFlags, flag)
	return a
}

// WithAlias sets the alias for the flag.
func (f *Flag) WithAlias(alias string) *Flag {
	f.alias = alias
	return f
}

// WithType sets the type for the flag.
func (f *Flag) WithType(flagType flagType) *Flag {
	f.defaultValue = defaultValues[flagType]
	f.flagType = flagType
	return f
}

// WithDefault sets the default value for the flag.
func (f *Flag) WithDefault(defaultValue any) *Flag {
	f.defaultValue = defaultValue
	return f
}

// WithDescription sets the description for the flag.
func (f *Flag) WithDescription(description string) *Flag {
	f.description = description
	return f
}

// WithEnum sets the allowed values for the flag.
func (f *Flag) WithEnum(values ...string) *Flag {
	f.allowedValues = values
	return f
}

// Name returns the name of the flag.
func (f *Flag) Name() string { return f.name }

// Alias returns the alias of the flag.
func (f *Flag) Alias() string { return f.alias }

// Type returns the type of the flag.
func (f *Flag) Type() flagType { return f.flagType }

// Default returns the default value of the flag.
func (f *Flag) Default() any { return f.defaultValue }

// Description returns the description of the flag.
func (f *Flag) Description() string { return f.description }

// AllowedValues returns the allowed values of the flag.
func (f *Flag) AllowedValues() []string { return f.allowedValues }
