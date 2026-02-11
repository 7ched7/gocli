package gocli

// Flags represents a parsed set of flag values for a command.
type Flags struct {
	pair map[string]any
}

// flagType represents the type used for a flag value.
type flagType int

const (
	String flagType = iota
	Int
	Float
	Bool
)

// String returns the value of the flag as a string.
// It panics if the flag is not a string or not present.
func (o *Flags) String(flag string) string {
	return o.pair[flag].(string)
}

// Int returns the value of the flag as an int.
// It panics if the flag is not an int or not present.
func (o *Flags) Int(flag string) int {
	return o.pair[flag].(int)
}

// Float returns the value of the flag as a float64.
// It panics if the flag is not a float or not present.
func (o *Flags) Float(flag string) float64 {
	return o.pair[flag].(float64)
}

// Bool returns the value of the flag as a bool.
// It panics if the flag is not a bool or not present.
func (o *Flags) Bool(flag string) bool {
	return o.pair[flag].(bool)
}
