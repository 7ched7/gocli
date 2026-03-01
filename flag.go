package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, expected value type, default value,
// and description for help menu.
type Flag[T any] struct {
	name         string
	alias        string
	value        FlagValue
	defaultValue FlagValue
	description  string
	helpType     string
	validator    func(ctx *Context, value T) error
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
	HelpType() string
	Validate(ctx *Context) error
}

func NewStringFlag(name string, defaultValue string) *Flag[string] {
	var value string
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: &value},
		defaultValue: &typeString{value: &defaultValue},
		helpType:     "string",
	}
}

func NewStringFlagVar(name string, value *string, defaultValue string) *Flag[string] {
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: value},
		defaultValue: &typeString{value: &defaultValue},
		helpType:     "string",
	}
}

func NewIntFlag(name string, defaultValue int) *Flag[int] {
	var value int
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: &value},
		defaultValue: &typeInt{value: &defaultValue},
		helpType:     "int",
	}
}

func NewIntFlagVar(name string, value *int, defaultValue int) *Flag[int] {
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: value},
		defaultValue: &typeInt{value: &defaultValue},
		helpType:     "int",
	}
}

func NewFloatFlag(name string, defaultValue float64) *Flag[float64] {
	var value float64
	return &Flag[float64]{
		name: name, value: &typeFloat{value: &value},
		defaultValue: &typeFloat{value: &defaultValue},
		helpType:     "float64",
	}
}

func NewFloatFlagVar(name string, value *float64, defaultValue float64) *Flag[float64] {
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat{value: value},
		defaultValue: &typeFloat{value: &defaultValue},
		helpType:     "float64",
	}
}

func NewBoolFlag(name string, defaultValue bool) *Flag[bool] {
	var value bool
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: &value},
		defaultValue: &typeBool{value: &defaultValue},
		helpType:     "bool",
	}
}

func NewBoolFlagVar(name string, value *bool, defaultValue bool) *Flag[bool] {
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: value},
		defaultValue: &typeBool{value: &defaultValue},
		helpType:     "bool",
	}
}

func NewStringSliceFlag(name string, defaultValue []string) *Flag[[]string] {
	var value []string
	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: &value},
		defaultValue: &typeStringSlice{value: &defaultValue},
		helpType:     "strings",
	}
}

func NewStringSliceFlagVar(name string, value *[]string, defaultValue []string) *Flag[[]string] {
	return &Flag[[]string]{
		name:         name,
		value:        &typeStringSlice{value: value},
		defaultValue: &typeStringSlice{value: &defaultValue},
		helpType:     "strings",
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

func (f *Flag[T]) WithHelpType(helpType string) *Flag[T] {
	f.helpType = helpType
	return f
}

func (f *Flag[T]) WithValidator(fn func(ctx *Context, value T) error) *Flag[T] {
	f.validator = fn
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

func (f *Flag[T]) HelpType() string { return f.helpType }

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
