package gocli

// Flag represents a single flag for the CLI.
// It includes flag name, alias, parsed value, description,
// metavariable, and an optional validator function.
type Flag[T any] struct {
	name         string
	alias        string
	value        FlagValue
	defaultValue any
	description  string
	metavar      string
	validator    func(ctx *Context, value T) error
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
	DefaultValue() any           // DefaultValue returns the default value of the flag.
	Metavar() string             // Metavar returns the metavariable of the flag.
	Validate(ctx *Context) error // Validate runs the flag validator function, if set.
}

// NewStringFlag creates a new string flag with the given name and default value.
func NewStringFlag(name string, defaultValue string) *Flag[string] {
	value := defaultValue
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: &value},
		defaultValue: defaultValue,
		metavar:      "STRING",
	}
}

// NewStringFlagVar creates a new string flag with the given name and provided variable.
func NewStringFlagVar(name string, variable *string) *Flag[string] {
	defaultValue := *variable
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: variable},
		defaultValue: defaultValue,
		metavar:      "STRING",
	}
}

// NewIntFlag creates a new int flag with the given name and default value.
func NewIntFlag(name string, defaultValue int) *Flag[int] {
	value := defaultValue
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: &value},
		defaultValue: defaultValue,
		metavar:      "INT",
	}
}

// NewIntFlagVar creates a new int flag with the given name and provided variable.
func NewIntFlagVar(name string, variable *int) *Flag[int] {
	defaultValue := *variable
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: variable},
		defaultValue: defaultValue,
		metavar:      "INT",
	}
}

// NewFloatFlag creates a new float64 flag with the given name and default value.
func NewFloatFlag(name string, defaultValue float64) *Flag[float64] {
	value := defaultValue
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat64{value: &value},
		defaultValue: defaultValue,
		metavar:      "FLOAT",
	}
}

// NewFloatFlagVar creates a new float64 flag with the given name and provided variable.
func NewFloatFlagVar(name string, variable *float64) *Flag[float64] {
	defaultValue := *variable
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat64{value: variable},
		defaultValue: defaultValue,
		metavar:      "FLOAT",
	}
}

// NewBoolFlag creates a new bool flag with the given name and default value.
func NewBoolFlag(name string, defaultValue bool) *Flag[bool] {
	value := defaultValue
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: &value},
		defaultValue: defaultValue,
		metavar:      "BOOL",
	}
}

// NewBoolFlagVar creates a new bool flag with the given name and provided variable.
func NewBoolFlagVar(name string, variable *bool) *Flag[bool] {
	defaultValue := *variable
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: variable},
		defaultValue: defaultValue,
		metavar:      "BOOL",
	}
}

// NewStringSliceFlag creates a new string slice flag with the given name and default value.
func NewStringSliceFlag(name string, defaultValue []string) *Flag[[]string] {
	var value []string

	if len(defaultValue) > 0 {
		value = make([]string, len(defaultValue))
		copy(value, defaultValue)
	}

	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: &value},
		defaultValue: defaultValue,
		metavar:      "STRINGS",
	}
}

// NewStringSliceFlagVar creates a new string slice flag with the given name and provided variable.
func NewStringSliceFlagVar(name string, variable *[]string) *Flag[[]string] {
	var defaultValue []string

	if len(*variable) > 0 {
		defaultValue = make([]string, len(*variable))
		copy(defaultValue, *variable)
	}

	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: variable},
		defaultValue: defaultValue,
		metavar:      "STRINGS",
	}
}

// NewCustomFlagVar creates a new custom flag with the given name and provided variable.
func NewCustomFlagVar(name string, variable FlagValue) *Flag[FlagValue] {
	return &Flag[FlagValue]{
		name:         name,
		value:        variable,
		defaultValue: variable.Get(),
	}
}

// AddFlag registers flags to the command.
func (c *Command) AddFlag(flags ...FlagInfo) *Command {
	for _, f := range flags {
		c.flags = append(c.flags, f)
	}
	return c
}

// AddGlobalFlag registers global flags to the application.
// Global flags apply to all commands.
func (a *App) AddGlobalFlag(flags ...FlagInfo) *App {
	a.root.AddFlag(flags...)
	return a
}

// WithAlias sets the alias for the flag.
func (f *Flag[T]) WithAlias(alias string) *Flag[T] {
	f.alias = alias
	return f
}

// WithDescription sets the description for the flag.
// This is shown in flags section within help menu.
func (f *Flag[T]) WithDescription(description string) *Flag[T] {
	f.description = description
	return f
}

// WithMetavar sets the metavariable for the flag.
// This is shown next to the flag and indicates the type of the flag value.
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

// DefaultValue returns the default value of the flag.
func (f *Flag[T]) DefaultValue() any { return f.defaultValue }

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
