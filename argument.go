package gocli

// Argument represents an argument for the CLI.
// It defines the allowed minimum and maximum numbers
// along with metadata and holds the actual values.
type Argument struct {
	name  string
	min   int
	max   int
	value *[]string
}

// ArgumentInfo provides access to argument metadata.
type ArgumentInfo interface {
	Name() string     // Name returns the name of the argument.
	Min() int         // Min returns the minimum number of arguments required.
	Max() int         // Max returns the maximum number of arguments required.
	First() string    // First returns the first value of the argument.
	Last() string     // Last returns the last value of the argument.
	All() []string    // All returns all values provided for the argument.
	IsRequired() bool // IsRequired returns whether the argument is required.
	IsVariadic() bool // IsVariadic returns whether the argument is variadic.

	set(value ...string)
}

// NewArgument creates a new argument with the given name.
func NewArgument(name string) *Argument {
	value := []string{}
	return &Argument{
		name:  name,
		min:   1,
		max:   1,
		value: &value,
	}
}

// NewArgumentVar creates a new argument with the given name and provided variable.
func NewArgumentVar(name string, variable *[]string) *Argument {
	return &Argument{
		name:  name,
		min:   1,
		max:   1,
		value: variable,
	}
}

// WithRange sets the minimum and maximum number for the argument.
func (a *Argument) WithRange(min, max int) *Argument {
	if min < 0 {
		return a
	}

	if max != -1 && min > max {
		return a
	}

	if min == 0 && max == 0 {
		return a
	}

	a.min = min
	a.max = max
	return a
}

// Name returns the name of the argument.
func (a *Argument) Name() string { return a.name }

// Min returns the minimum number of arguments required.
func (a *Argument) Min() int { return a.min }

// Max returns the maximum number of arguments required.
func (a *Argument) Max() int { return a.max }

// First returns the first value of the argument.
func (a *Argument) First() string {
	if len(*a.value) == 0 {
		return ""
	}
	return (*a.value)[0]
}

// Last returns the last value of the argument.
func (a *Argument) Last() string {
	if len(*a.value) == 0 {
		return ""
	}
	return (*a.value)[len(*a.value)-1]
}

// All returns all values provided for the argument.
func (a *Argument) All() []string {
	value := make([]string, len(*a.value))
	copy(value, *a.value)
	return value
}

// IsRequired returns whether the argument is required.
func (a *Argument) IsRequired() bool { return a.min >= 1 }

// IsVariadic returns whether the argument is variadic.
func (a *Argument) IsVariadic() bool { return a.max == -1 || a.max > 1 }

func (a *Argument) set(value ...string) { *a.value = append(*a.value, value...) }
