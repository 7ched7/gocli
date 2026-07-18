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
	IsRequired() bool // IsRequired returns whether the argument is required.
	IsVariadic() bool // IsVariadic returns whether the argument is variadic.
	ArgumentGetter

	set(value ...string)
}

// ArgumentGetter defines an interface for accessing argument values.
type ArgumentGetter interface {
	Get(index int) string // Get returns the value of the argument with the given index.
	All() []string        // All returns all values provided for the argument.
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

// IsRequired returns whether the argument is required.
func (a *Argument) IsRequired() bool { return a.min >= 1 }

// IsVariadic returns whether the argument is variadic.
func (a *Argument) IsVariadic() bool { return a.max == -1 || a.max > 1 }

// Get returns the value of the argument with the given index.
func (a *Argument) Get(index int) string {
	len := len(*a.value)

	if index < 0 {
		index = len + index
	}

	if index < 0 || index >= len {
		return ""
	}
	return (*a.value)[index]
}

// All returns all values provided for the argument.
func (a *Argument) All() []string {
	return *a.value
}

func (a *Argument) set(value ...string) { *a.value = append(*a.value, value...) }
