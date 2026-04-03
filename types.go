package gocli

import (
	"fmt"
	"strconv"
	"strings"
)

// Context holds positional arguments
// and parsed flags for a command execution.
type Context struct {
	app     *App
	command *Command
	args    []string
	flags   map[string]FlagValue
}

// App returns the application instance.
func (c *Context) App() *App { return c.app }

// Command returns the executed command.
func (c *Context) Command() *Command { return c.command }

// Args returns the positional arguments passed to the command.
func (c *Context) Args() []string { return c.args }

// Flags returns all parsed flags as key-value pairs.
func (c *Context) Flags() map[string]FlagValue { return c.flags }

// Lookup returns the FlagValue interface of the flag with the given name.
func (c *Context) Lookup(name string) FlagValue { return c.flags[name] }

// String returns the named flag as string.
// It panics if not found or invalid type.
func (c *Context) String(name string) string {
	return c.Lookup(name).Get().(string)
}

// Int returns the named flag as int.
// It panics if not found or invalid type.
func (c *Context) Int(name string) int {
	return c.Lookup(name).Get().(int)
}

// Float64 returns the named flag as float64.
// It panics if not found or invalid type.
func (c *Context) Float64(name string) float64 {
	return c.Lookup(name).Get().(float64)
}

// Bool returns the named flag as bool.
// It panics if not found or invalid type.
func (c *Context) Bool(name string) bool {
	return c.Lookup(name).Get().(bool)
}

// StringSlice returns the named flag as []string.
// It panics if not found or invalid type.
func (c *Context) StringSlice(name string) []string {
	return c.Lookup(name).Get().([]string)
}

/*
------------------------------
STRING
------------------------------
*/
type typeString struct {
	value *string
}

func (s *typeString) Set(value string) error {
	*s.value = value
	return nil
}
func (s *typeString) Get() any       { return *s.value }
func (s *typeString) String() string { return *s.value }

/*
------------------------------
INT
------------------------------
*/
type typeInt struct {
	value *int
}

func (i *typeInt) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be an integer.\n", value)
	}
	*i.value = v
	return nil
}
func (i *typeInt) Get() any       { return int(*i.value) }
func (i *typeInt) String() string { return strconv.FormatInt(int64(*i.value), 10) }

/*
------------------------------
FLOAT64
------------------------------
*/
type typeFloat64 struct {
	value *float64
}

func (f *typeFloat64) Set(value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be a float.\n", value)
	}
	*f.value = v
	return nil
}
func (f *typeFloat64) Get() any       { return float64(*f.value) }
func (f *typeFloat64) String() string { return fmt.Sprintf("%v", *f.value) }

/*
------------------------------
BOOL
------------------------------
*/
type typeBool struct {
	value *bool
}

func (b *typeBool) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be a bool.\n", value)
	}
	*b.value = v
	return nil
}
func (b *typeBool) Get() any       { return bool(*b.value) }
func (b *typeBool) String() string { return fmt.Sprintf("%v", *b.value) }

/*
------------------------------
[]STRING
------------------------------
*/
type typeStringSlice struct {
	value *[]string
	isSet bool
}

func (ss *typeStringSlice) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		if !ss.isSet {
			*ss.value = []string{v}
			ss.isSet = true
		} else {
			*ss.value = append(*ss.value, v)
		}
	}
	return nil
}
func (ss *typeStringSlice) Get() any {
	slc := make([]string, 0)
	for _, v := range *ss.value {
		slc = append(slc, v)
	}
	return slc
}
func (ss *typeStringSlice) String() string { return strings.Join(*ss.value, ",") }
