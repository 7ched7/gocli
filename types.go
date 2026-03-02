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

var zeroValues = map[string]string{
	"string":  "",
	"int":     "0",
	"float64": "0",
	"bool":    "false",
	"strings": "",
}

// TypeString implements FlagValue for string flags.
type TypeString struct {
	value *string
}

func (s *TypeString) Set(value string) error {
	*s.value = value
	return nil
}
func (s *TypeString) Get() any       { return *s.value }
func (s *TypeString) String() string { return *s.value }

// TypeInt implements FlagValue for integer flags.
type TypeInt struct {
	value *int
}

func (i *TypeInt) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be an integer.\n", value)
	}
	*i.value = v
	return nil
}
func (i *TypeInt) Get() any       { return int(*i.value) }
func (i *TypeInt) String() string { return strconv.FormatInt(int64(*i.value), 10) }

// TypeFloat implements FlagValue for float64 flags.
type TypeFloat struct {
	value *float64
}

func (f *TypeFloat) Set(value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be a float.\n", value)
	}
	*f.value = v
	return nil
}
func (f *TypeFloat) Get() any       { return float64(*f.value) }
func (f *TypeFloat) String() string { return fmt.Sprintf("%v", *f.value) }

// TypeBool implements FlagValue for boolean flags.
type TypeBool struct {
	value *bool
}

func (b *TypeBool) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be a bool.\n", value)
	}
	*b.value = v
	return nil
}
func (b *TypeBool) Get() any       { return bool(*b.value) }
func (b *TypeBool) String() string { return fmt.Sprintf("%v", *b.value) }

// TypeStringSlice implements FlagValue for string slice flags.
type TypeStringSlice struct {
	value *[]string
}

func (ss *TypeStringSlice) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		*ss.value = append(*ss.value, v)
	}
	return nil
}
func (ss *TypeStringSlice) Get() any {
	slc := make([]string, 0)
	for _, v := range *ss.value {
		slc = append(slc, v)
	}
	return slc
}

func (ss *TypeStringSlice) String() string { return strings.Join(*ss.value, ",") }

// App returns the application instance.
func (c *Context) App() *App { return c.app }

// Command returns the executed command.
func (c *Context) Command() *Command { return c.command }

// Args returns the positional arguments passed to the command.
func (c *Context) Args() []string { return c.args }

// Flags returns all parsed flags as a key-value pairs.
func (c *Context) Flags() map[string]FlagValue { return c.flags }

// Flag returns the value of a flag by name.
func (c *Context) Flag(value string) any { return c.flags[value].Get() }

// String returns the value of the flag as a string.
// It panics if the flag is not a string or not present.
func (c *Context) String(flag string) string {
	return c.flags[flag].Get().(string)
}

// Int returns the value of the flag as an int.
// It panics if the flag is not an int or not present.
func (c *Context) Int(flag string) int {
	return c.flags[flag].Get().(int)
}

// Float returns the value of the flag as a float64.
// It panics if the flag is not a float64 or not present.
func (c *Context) Float(flag string) float64 {
	return c.flags[flag].Get().(float64)
}

// Bool returns the value of the flag as a bool.
// It panics if the flag is not a bool or not present.
func (c *Context) Bool(flag string) bool {
	return c.flags[flag].Get().(bool)
}

// StringSlice returns the value of the flag as a slice of strings.
// It panics if the flag is not a slice of strings or not present.
func (c *Context) StringSlice(flag string) []string {
	return c.flags[flag].Get().([]string)
}
