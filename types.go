package gocli

import "strconv"

// Flags represents a parsed set of flag values for a command.
type Flags struct {
	pair map[string]any
}

type flagType int

const (
	String flagType = iota
	Int
	Float
	Bool
	StringSlice
	IntSlice
	FloatSlice
	BoolSlice
)

var defaultValues = map[flagType]any{
	String:      "",
	Int:         0,
	Float:       0.0,
	Bool:        false,
	StringSlice: []string{},
	IntSlice:    []int{},
	FloatSlice:  []float64{},
	BoolSlice:   []bool{},
}

var flagTypeHandlerMap = map[flagType]func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int{
	String: func(_ *App, _ *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		flags.pair[matchedFlag.name] = flagValue
		return StateContinue
	},

	Int: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		i, code := a.parseInt(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = i
		return StateContinue
	},

	Float: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		f, code := a.parseFloat(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = f
		return StateContinue
	},

	Bool: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		b, code := a.parseBool(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = b
		return StateContinue
	},

	StringSlice: func(_ *App, _ *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		flags.pair[matchedFlag.name] = append(flags.pair[matchedFlag.name].([]string), flagValue)
		return StateContinue
	},

	IntSlice: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		i, code := a.parseInt(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = append(flags.pair[matchedFlag.name].([]int), i)
		return StateContinue
	},

	FloatSlice: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		f, code := a.parseFloat(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = append(flags.pair[matchedFlag.name].([]float64), f)
		return StateContinue
	},

	BoolSlice: func(a *App, cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
		b, code := a.parseBool(cmd, flagValue)
		if code != StateContinue {
			return code
		}
		flags.pair[matchedFlag.name] = append(flags.pair[matchedFlag.name].([]bool), b)
		return StateContinue
	},
}

func (a *App) parseInt(cmd *Command, flagValue string) (int, int) {
	parsed, err := strconv.Atoi(flagValue)
	if err != nil {
		return 0, a.stop(ErrInvalidIntValue, cmd, map[string]any{
			"value": flagValue,
		})
	}
	return parsed, StateContinue
}

func (a *App) parseFloat(cmd *Command, flagValue string) (float64, int) {
	parsed, err := strconv.ParseFloat(flagValue, 64)
	if err != nil {
		return 0.0, a.stop(ErrInvalidFloatValue, cmd, map[string]any{
			"value": flagValue,
		})
	}
	return parsed, StateContinue
}

func (a *App) parseBool(cmd *Command, flagValue string) (bool, int) {
	parsed, err := strconv.ParseBool(flagValue)
	if err != nil {
		return false, a.stop(ErrInvalidBoolValue, cmd, map[string]any{
			"value": flagValue,
		})
	}
	return parsed, StateContinue
}

// String returns the value of the flag as a string.
// It panics if the flag is not a string or not present.
func (f *Flags) String(flag string) string {
	return f.pair[flag].(string)
}

// Int returns the value of the flag as an int.
// It panics if the flag is not an int or not present.
func (f *Flags) Int(flag string) int {
	return f.pair[flag].(int)
}

// Float returns the value of the flag as a float64.
// It panics if the flag is not a float64 or not present.
func (f *Flags) Float(flag string) float64 {
	return f.pair[flag].(float64)
}

// Bool returns the value of the flag as a bool.
// It panics if the flag is not a bool or not present.
func (f *Flags) Bool(flag string) bool {
	return f.pair[flag].(bool)
}

// StringSlice returns the value of the flag as a slice of strings.
// It panics if the flag is not a slice of strings or not present.
func (f *Flags) StringSlice(flag string) []string {
	return f.pair[flag].([]string)
}

// IntSlice returns the value of the flag as a slice of ints.
// It panics if the flag is not a slice of ints or not present.
func (f *Flags) IntSlice(flag string) []int {
	return f.pair[flag].([]int)
}

// FloatSlice returns the value of the flag as a slice of float64s.
// It panics if the flag is not a slice of float64s or not present.
func (f *Flags) FloatSlice(flag string) []float64 {
	return f.pair[flag].([]float64)
}

// BoolSlice returns the value of the flag as a slice of bools.
// It panics if the flag is not a slice of bools or not present.
func (f *Flags) BoolSlice(flag string) []bool {
	return f.pair[flag].([]bool)
}
