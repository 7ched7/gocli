package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, parsed value, description,
// metavariable, and an optional validator function.
type Flag[T any] struct {
	name        string
	alias       string
	value       FlagValue
	description string
	metavar     string
	validator   func(ctx *Context, value T) error
}

// FlagValue defines an interface for all flag values.
type FlagValue interface {
	Set(value string) error // Set parses and assigns the given string to the underlying typed value.
	Get() any               // Get returns the underlying typed value.
	String() string         // String returns the string representation of the value.
}

// FlagInfo provides access to flag metadata and behaviour.
type FlagInfo interface {
	Name() string                // Name returns the name of the flag.
	Alias() string               // Alias returns the optional alias of the flag.
	Value() FlagValue            // Value returns the parsed value of the flag.
	Description() string         // Description returns the description of the flag.
	Metavar() string             // Metavar returns the metavariable of the flag.
	Validate(ctx *Context) error // Validate runs the flag validator function, if set.
}

// NewStringFlag creates a new string flag with the given name.
func NewStringFlag(name string) *Flag[string] {
	var value string
	return &Flag[string]{
		name:    name,
		value:   &TypeString{value: &value},
		metavar: "STRING",
	}
}

// NewStringFlagVar creates a new string flag using the provided variable.
func NewStringFlagVar(name string, variable *string) *Flag[string] {
	return &Flag[string]{
		name:    name,
		value:   &TypeString{value: variable},
		metavar: "STRING",
	}
}

// NewIntFlag creates a new int flag with the given name.
func NewIntFlag(name string) *Flag[int] {
	var value int
	return &Flag[int]{
		name:    name,
		value:   &TypeInt{value: &value},
		metavar: "INT",
	}
}

// NewIntFlagVar creates a new int flag using the provided variable.
func NewIntFlagVar(name string, variable *int) *Flag[int] {
	return &Flag[int]{
		name:    name,
		value:   &TypeInt{value: variable},
		metavar: "INT",
	}
}

// NewFloatFlag creates a new float64 flag with the given name.
func NewFloatFlag(name string) *Flag[float64] {
	var value float64
	return &Flag[float64]{
		name:    name,
		value:   &TypeFloat{value: &value},
		metavar: "FLOAT",
	}
}

// NewFloatFlagVar creates a new float64 flag using the provided variable.
func NewFloatFlagVar(name string, variable *float64) *Flag[float64] {
	return &Flag[float64]{
		name:    name,
		value:   &TypeFloat{value: variable},
		metavar: "FLOAT",
	}
}

// NewBoolFlag creates a new bool flag with the given name.
func NewBoolFlag(name string) *Flag[bool] {
	var value bool
	return &Flag[bool]{
		name:    name,
		value:   &TypeBool{value: &value},
		metavar: "BOOL",
	}
}

// NewBoolFlagVar creates a new bool flag using the provided variable.
func NewBoolFlagVar(name string, variable *bool) *Flag[bool] {
	return &Flag[bool]{
		name:    name,
		value:   &TypeBool{value: variable},
		metavar: "BOOL",
	}
}

// NewStringSliceFlag creates a new string slice flag with the given name.
func NewStringSliceFlag(name string) *Flag[[]string] {
	var value []string
	return &Flag[[]string]{
		name:    name,
		value:   &TypeStringSlice{value: &value},
		metavar: "STRINGS",
	}
}

// NewStringSliceFlagVar creates a new string slice flag using the provided variable.
func NewStringSliceFlagVar(name string, variable *[]string) *Flag[[]string] {
	return &Flag[[]string]{
		name:    name,
		value:   &TypeStringSlice{value: variable},
		metavar: "STRINGS",
	}
}

// NewCustomFlagVar creates a new flag using custom FlagValue implementations.
func NewCustomFlagVar(name string, variable FlagValue) *Flag[FlagValue] {
	return &Flag[FlagValue]{
		name:  name,
		value: variable,
	}
}

// AddFlag registers a flag to the command.
func (c *Command) AddFlag(flag FlagInfo) *Command {
	c.flags = append(c.flags, flag)
	return c
}

// AddGlobalFlag registers a global flag to the application.
// Global flags apply to all commands.
func (a *App) AddGlobalFlag(flag FlagInfo) *App {
	a.globalFlags = append(a.globalFlags, flag)
	return a
}

// WithAlias sets the alias for the flag.
func (f *Flag[T]) WithAlias(alias string) *Flag[T] {
	f.alias = alias
	return f
}

// WithDescription sets the description for the flag.
func (f *Flag[T]) WithDescription(description string) *Flag[T] {
	f.description = description
	return f
}

// WithMetavar sets the metavariable for the flag.
func (f *Flag[T]) WithMetavar(metavar string) *Flag[T] {
	f.metavar = metavar
	return f
}

// WithValidator registers a validation function for the flag.
func (f *Flag[T]) WithValidator(fn func(ctx *Context, value T) error) *Flag[T] {
	f.validator = fn
	return f
}

// Name returns the name of the flag.
func (f *Flag[T]) Name() string { return f.name }

// Alias returns the alias of the flag.
// If not set, it returns an empty string.
func (f *Flag[T]) Alias() string { return f.alias }

// Value returns the value of the flag.
func (f *Flag[T]) Value() FlagValue { return f.value }

// Description returns the description of the flag.
// If not set, it returns an empty string.
func (f *Flag[T]) Description() string { return f.description }

// Metavar returns the metavariable of the flag.
func (f *Flag[T]) Metavar() string { return f.metavar }

// Validate runs the flag validator function, if set.
func (f *Flag[T]) Validate(ctx *Context) error {
	if f.validator == nil {
		return nil
	}

	switch val := f.value.(type) {
	case T:
		return f.validator(ctx, val)
	case FlagValue:
		return f.validator(ctx, val.Get().(T))
	default:
		panic("invalid type")
	}
}
