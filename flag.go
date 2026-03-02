package gocli

// Flag represents a single flag for a command.
// It includes flag name, alias, parsed value, default value,
// description, help type, and an optional validator function.
type Flag[T any] struct {
	name         string
	alias        string
	value        FlagValue
	defaultValue FlagValue
	description  string
	helpType     string
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
	DefaultValue() FlagValue     // DefaultValue returns the default value of the flag.
	Description() string         // Description returns the description of the flag.
	HelpType() string            // HelpType returns the help type of the flag.
	Validate(ctx *Context) error // Validate runs the flag validator function, if set.
}

// NewStringFlag creates a new string flag with the given name and default value.
func NewStringFlag(name string, defaultValue string) *Flag[string] {
	var value string
	return &Flag[string]{
		name:         name,
		value:        &TypeString{value: &value},
		defaultValue: &TypeString{value: &defaultValue},
		helpType:     "string",
	}
}

// NewStringFlagVar creates a new string flag that directly binds its value.
func NewStringFlagVar(name string, value *string, defaultValue string) *Flag[string] {
	return &Flag[string]{
		name:         name,
		value:        &TypeString{value: value},
		defaultValue: &TypeString{value: &defaultValue},
		helpType:     "string",
	}
}

// NewIntFlag creates a new int flag with the given name and default value.
func NewIntFlag(name string, defaultValue int) *Flag[int] {
	var value int
	return &Flag[int]{
		name:         name,
		value:        &TypeInt{value: &value},
		defaultValue: &TypeInt{value: &defaultValue},
		helpType:     "int",
	}
}

// NewIntFlagVar creates a new int flag that directly binds its value.
func NewIntFlagVar(name string, value *int, defaultValue int) *Flag[int] {
	return &Flag[int]{
		name:         name,
		value:        &TypeInt{value: value},
		defaultValue: &TypeInt{value: &defaultValue},
		helpType:     "int",
	}
}

// NewFloatFlag creates a new float64 flag with the given name and default value.
func NewFloatFlag(name string, defaultValue float64) *Flag[float64] {
	var value float64
	return &Flag[float64]{
		name: name, value: &TypeFloat{value: &value},
		defaultValue: &TypeFloat{value: &defaultValue},
		helpType:     "float64",
	}
}

// NewFloatFlagVar creates a new float64 flag that directly binds its value.
func NewFloatFlagVar(name string, value *float64, defaultValue float64) *Flag[float64] {
	return &Flag[float64]{
		name:         name,
		value:        &TypeFloat{value: value},
		defaultValue: &TypeFloat{value: &defaultValue},
		helpType:     "float64",
	}
}

// NewBoolFlag creates a new bool flag with the given name and default value.
func NewBoolFlag(name string, defaultValue bool) *Flag[bool] {
	var value bool
	return &Flag[bool]{
		name:         name,
		value:        &TypeBool{value: &value},
		defaultValue: &TypeBool{value: &defaultValue},
		helpType:     "bool",
	}
}

// NewBoolFlagVar creates a new bool flag that directly binds its value.
func NewBoolFlagVar(name string, value *bool, defaultValue bool) *Flag[bool] {
	return &Flag[bool]{
		name:         name,
		value:        &TypeBool{value: value},
		defaultValue: &TypeBool{value: &defaultValue},
		helpType:     "bool",
	}
}

// NewStringSliceFlag creates a new string slice with the given name and default value.
func NewStringSliceFlag(name string, defaultValue []string) *Flag[[]string] {
	var value []string
	return &Flag[[]string]{
		name:         name,
		value:        &TypeStringSlice{value: &value},
		defaultValue: &TypeStringSlice{value: &defaultValue},
		helpType:     "strings",
	}
}

// NewStringSliceFlagVar creates a new string slice flag that directly binds its value.
func NewStringSliceFlagVar(name string, value *[]string, defaultValue []string) *Flag[[]string] {
	return &Flag[[]string]{
		name:         name,
		value:        &TypeStringSlice{value: value},
		defaultValue: &TypeStringSlice{value: &defaultValue},
		helpType:     "strings",
	}
}

// NewCustomFlagVar creates a new flag using custom FlagValue implementations.
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

// WithHelpType sets the help type shown in flags within help menu.
func (f *Flag[T]) WithHelpType(helpType string) *Flag[T] {
	f.helpType = helpType
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

// DefaultValue returns the default value of the flag.
func (f *Flag[T]) DefaultValue() FlagValue { return f.defaultValue }

// Description returns the description of the flag.
// If not set, it returns an empty string.
func (f *Flag[T]) Description() string { return f.description }

// HelpType returns the help type of the flag.
func (f *Flag[T]) HelpType() string { return f.helpType }

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
