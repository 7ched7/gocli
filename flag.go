package gocli

import "fmt"

type flagRole int

const (
	flagStandard flagRole = iota
	flagHelp
	flagVersion
)

// Flag represents a single flag for the CLI.
// It handles parsing, validation, and metadata for the flag.
type Flag[T any] struct {
	name         string
	alias        string
	value        FlagValue
	defaultValue any
	description  string
	placeholder  string
	isRequired   bool
	validators   []func(ctx *Context, value T) error
	isSet        bool
	r            flagRole
}

// FlagInfo provides access to flag metadata.
type FlagInfo interface {
	Name() string                                      // Name returns the name of the flag.
	Alias() string                                     // Alias returns the optional alias of the flag.
	Value() FlagValue                                  // Value returns the parsed value of the flag.
	DefaultValue() any                                 // DefaultValue returns the default value of the flag.
	Description() string                               // Description returns the description of the flag.
	Placeholder() string                               // Placeholder returns the placeholder of the flag.
	IsRequired() bool                                  // IsRequired returns whether the flag is required.
	Validators() []func(ctx *Context, value any) error // Validators returns a list of validation functions defined for the flag.
	Validate(ctx *Context) error                       // Validate runs the validator functions defined for the flag.
	IsSet() bool                                       // IsSet returns whether the flag is set.
	FlagValueGetter

	set()
	setRole(role flagRole)
	role() flagRole
}

// FlagValue defines an interface for all flag values.
type FlagValue interface {
	Set(value string) error // Set parses and assigns the given string to the underlying typed value.
	Get() any               // Get returns the underlying typed value.
	String() string         // String returns the string representation of the value.
}

// NoArgFlag defines an interface for flags that can be used without an argument.
type NoArgFlag interface {
	SetNoArg() error // SetNoArg sets the value of the flag when no argument is provided.
}

// FlagValueGetter defines an interface for retrieving typed flag values.
type FlagValueGetter interface {
	String() string        // String returns the value of the flag as string.
	Int() int              // Int returns the value of the flag as int.
	Bool() bool            // Bool returns the value of the flag as bool.
	Float64() float64      // Float64 returns the value of the flag as float64.
	StringSlice() []string // StringSlice returns the value of the flag as []string.
}

// NewStringFlag creates a new string flag with the given name and default value.
func NewStringFlag(name string, defaultValue string) *Flag[string] {
	value := defaultValue
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: &value},
		defaultValue: defaultValue,
		placeholder:  "STRING",
	}
}

// NewStringFlagVar creates a new string flag with the given name and provided variable.
func NewStringFlagVar(name string, variable *string) *Flag[string] {
	defaultValue := *variable
	return &Flag[string]{
		name:         name,
		value:        &typeString{value: variable},
		defaultValue: defaultValue,
		placeholder:  "STRING",
	}
}

// NewIntFlag creates a new int flag with the given name and default value.
func NewIntFlag(name string, defaultValue int) *Flag[int] {
	value := defaultValue
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: &value},
		defaultValue: defaultValue,
		placeholder:  "INT",
	}
}

// NewIntFlagVar creates a new int flag with the given name and provided variable.
func NewIntFlagVar(name string, variable *int) *Flag[int] {
	defaultValue := *variable
	return &Flag[int]{
		name:         name,
		value:        &typeInt{value: variable},
		defaultValue: defaultValue,
		placeholder:  "INT",
	}
}

// NewFloatFlag creates a new float64 flag with the given name and default value.
func NewFloatFlag(name string, defaultValue float64) *Flag[float64] {
	value := defaultValue
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat64{value: &value},
		defaultValue: defaultValue,
		placeholder:  "FLOAT",
	}
}

// NewFloatFlagVar creates a new float64 flag with the given name and provided variable.
func NewFloatFlagVar(name string, variable *float64) *Flag[float64] {
	defaultValue := *variable
	return &Flag[float64]{
		name:         name,
		value:        &typeFloat64{value: variable},
		defaultValue: defaultValue,
		placeholder:  "FLOAT",
	}
}

// NewBoolFlag creates a new bool flag with the given name and default value.
func NewBoolFlag(name string, defaultValue bool) *Flag[bool] {
	value := defaultValue
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: &value},
		defaultValue: defaultValue,
	}
}

// NewBoolFlagVar creates a new bool flag with the given name and provided variable.
func NewBoolFlagVar(name string, variable *bool) *Flag[bool] {
	defaultValue := *variable
	return &Flag[bool]{
		name:         name,
		value:        &typeBool{value: variable},
		defaultValue: defaultValue,
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
		placeholder:  "STRINGS",
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
		placeholder:  "STRINGS",
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

// WithPlaceholder sets the placeholder for the flag.
// This is shown next to the flag and indicates the type of the flag value.
func (f *Flag[T]) WithPlaceholder(placeholder string) *Flag[T] {
	f.placeholder = placeholder
	return f
}

// WithRequired sets the flag to required.
func (f *Flag[T]) WithRequired() *Flag[T] {
	f.isRequired = true
	return f
}

// WithValidator registers validation functions to the flag.
func (f *Flag[T]) WithValidator(fn ...func(ctx *Context, value T) error) *Flag[T] {
	for _, v := range fn {
		f.validators = append(f.validators, v)
	}
	return f
}

// Name returns the name of the flag.
func (f *Flag[T]) Name() string { return f.name }

// Alias returns the alias of the flag.
func (f *Flag[T]) Alias() string { return f.alias }

// Value returns the value of the flag.
func (f *Flag[T]) Value() FlagValue { return f.value }

// DefaultValue returns the default value of the flag.
func (f *Flag[T]) DefaultValue() any { return f.defaultValue }

// Description returns the description of the flag.
func (f *Flag[T]) Description() string { return f.description }

// Placeholder returns the placeholder of the flag.
func (f *Flag[T]) Placeholder() string { return f.placeholder }

// IsRequired returns whether the flag is required.
func (f *Flag[T]) IsRequired() bool { return f.isRequired }

// Validators returns a list of validation functions defined for the flag.
func (f *Flag[T]) Validators() []func(ctx *Context, value any) error {
	out := make([]func(ctx *Context, value any) error, 0, len(f.validators))
	for _, v := range f.validators {
		vv := v
		out = append(out, func(ctx *Context, value any) error {
			return vv(ctx, value.(T))
		})
	}
	return out
}

// Validate runs the validator functions defined for the flag.
func (f *Flag[T]) Validate(ctx *Context) error {
	switch val := f.value.(type) {
	case T:
		for _, v := range f.validators {
			if err := v(ctx, val); err != nil {
				return err
			}
		}
		return nil
	case FlagValue:
		got := val.Get().(T)
		for _, v := range f.validators {
			if err := v(ctx, got); err != nil {
				return err
			}
		}
		return nil
	default:
		panic("invalid type")
	}
}

// IsSet returns whether the flag is set.
func (f *Flag[T]) IsSet() bool { return f.isSet }

// String returns the value of the flag as string.
func (f *Flag[T]) String() string {
	return fmt.Sprintf("%v", f.Value().Get())
}

// Int returns the value of the flag as int.
func (f *Flag[T]) Int() int {
	return f.Value().Get().(int)
}

// Bool returns the value of the flag as bool.
func (f *Flag[T]) Bool() bool {
	return f.Value().Get().(bool)
}

// Float64 returns the value of the flag as float64.
func (f *Flag[T]) Float64() float64 {
	return f.Value().Get().(float64)
}

// StringSlice returns the value of the flag as []string.
func (f *Flag[T]) StringSlice() []string {
	return f.Value().Get().([]string)
}

func (f *Flag[T]) set()               { f.isSet = true }
func (f *Flag[T]) setRole(r flagRole) { f.r = r }
func (f *Flag[T]) role() flagRole     { return f.r }
