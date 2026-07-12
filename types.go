package gocli

import (
	"fmt"
	"strconv"
	"strings"
)

// Context represents the runtime context of a command execution.
// It includes arguments, parsed flags, and references to the app and command.
type Context struct {
	app     AppInfo
	command CommandInfo
	args    map[string]ArgumentInfo
	flags   map[string]FlagInfo
}

// App returns the application instance.
func (c *Context) App() AppInfo { return c.app }

// Command returns the executed command.
func (c *Context) Command() CommandInfo { return c.command }

// Args returns all arguments as key-value pairs.
func (c *Context) Args() map[string]ArgumentInfo { return c.args }

// Arg returns the argument with the given name.
func (c *Context) Arg(name string) ArgumentInfo { return c.args[name] }

// Flags returns all parsed flags as key-value pairs.
func (c *Context) Flags() map[string]FlagInfo { return c.flags }

// Flag returns the flag with the given name.
func (c *Context) Flag(name string) FlagInfo { return c.flags[name] }

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
		return Exit(exitUsage, "")
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
		return Exit(exitUsage, "")
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
		return Exit(exitUsage, "")
	}
	*b.value = v
	return nil
}
func (b *typeBool) SetNoArg() error {
	*b.value = true
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
