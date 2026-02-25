package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, expected value type, default value,
// and description for help menu.
type Flag[T any] struct {
	name          string
	alias         string
	value         FlagValue
	defaultValue  FlagValue
	description   string
	allowedValues []any
}

// FlagValue defines a value that can be set from a string
// and represented as a string.
type FlagValue interface {
	Set(value string) error
	Get() any
	String() string
}

// FlagInfo provides access to flag metadata.
type FlagInfo interface {
	Name() string
	Alias() string
	Value() FlagValue
	DefaultValue() FlagValue
	Description() string
	AllowedValues() []any
}

func NewStringFlag(name string, defaultValue string) *Flag[string] {
	var value string
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: &value},
		defaultValue: &typeString{value: &defaultValue},
	}
}

func NewStringFlagVar(name string, value *string, defaultValue string) *Flag[string] {
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: value},
		defaultValue: &typeString{value: &defaultValue},
	}
}

func NewIntFlag(name string, defaultValue int) *Flag[int] {
	var value int
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: &value},
		defaultValue: &typeInt{value: &defaultValue},
	}
}

func NewIntFlagVar(name string, value *int, defaultValue int) *Flag[int] {
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: value},
		defaultValue: &typeInt{value: &defaultValue},
	}
}

func NewFloatFlag(name string, defaultValue float64) *Flag[float64] {
	var value float64
	return &Flag[float64]{
		name: name, value: &typeFloat{value: &value},
		defaultValue: &typeFloat{value: &defaultValue},
	}
}

func NewFloatFlagVar(name string, value *float64, defaultValue float64) *Flag[float64] {
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat{value: value},
		defaultValue: &typeFloat{value: &defaultValue},
	}
}

func NewBoolFlag(name string, defaultValue bool) *Flag[bool] {
	var value bool
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: &value},
		defaultValue: &typeBool{value: &defaultValue},
	}
}

func NewBoolFlagVar(name string, value *bool, defaultValue bool) *Flag[bool] {
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: value},
		defaultValue: &typeBool{value: &defaultValue},
	}
}

func NewStringSliceFlag(name string, defaultValue []string) *Flag[[]string] {
	var value []string
	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: &value},
		defaultValue: &typeStringSlice{value: &defaultValue},
	}
}

func NewStringSliceFlagVar(name string, value *[]string, defaultValue []string) *Flag[[]string] {
	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: value},
		defaultValue: &typeStringSlice{value: &defaultValue},
	}
}

func NewCustomFlagVar(name string, value FlagValue, defaultValue FlagValue) *Flag[FlagValue] {
	return &Flag[FlagValue]{
		name:         name,
		value:        value,
		defaultValue: defaultValue,
	}
}

// AddFlag registers a flag to the command.
func (c *Command) AddFlag(flag FlagInfo) *Command {
	c.flags = append(c.flags, flag)
	return c
}

// AddGlobalFlag registers a global flag to the application.
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

// WithEnum sets the allowed values for the flag.
func (f *Flag[T]) WithEnum(values ...any) *Flag[T] {
	f.allowedValues = values
	return f
}

// Name returns the name of the flag.
func (f *Flag[T]) Name() string { return f.name }

// Alias returns the alias of the flag.
func (f *Flag[T]) Alias() string { return f.alias }

// Value returns the value of the flag.
func (f *Flag[T]) Value() FlagValue { return f.value }

// DefaultValue returns the default value of the flag.
func (f *Flag[T]) DefaultValue() FlagValue { return f.defaultValue }

// Description returns the description of the flag.
func (f *Flag[T]) Description() string { return f.description }

// AllowedValues returns the allowed values of the flag.
func (f *Flag[T]) AllowedValues() []any { return f.allowedValues }
