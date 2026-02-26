package gocli

import (
	"fmt"
	"strconv"
	"strings"
)

// Context holds parsed positional arguments
// and flags for a command execution.
type Context struct {
	app     *App
	command *Command
	args    []string
	flags   map[string]FlagValue
}

type typeString struct {
	value *string
}

func (s *typeString) Set(value string) error {
	*s.value = value
	return nil
}
func (s *typeString) Get() any       { return *s.value }
func (s *typeString) String() string { return *s.value }

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

type typeFloat struct {
	value *float64
}

func (f *typeFloat) Set(value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("error: invalid value '%v': must be a float.\n", value)
	}
	*f.value = v
	return nil
}
func (f *typeFloat) Get() any       { return float64(*f.value) }
func (f *typeFloat) String() string { return fmt.Sprintf("%v", *f.value) }

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

type typeStringSlice struct {
	value *[]string
}

func (ss *typeStringSlice) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		*ss.value = append(*ss.value, v)
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

func (c *Context) App() *App { return c.app }

func (c *Context) Command() *Command { return c.command }

func (c *Context) Args() []string { return c.args }

func (c *Context) Flags() map[string]FlagValue { return c.flags }

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
