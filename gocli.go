package gocli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type App struct {
	name        string                                            // Application name
	version     string                                            // Optional app version
	description string                                            // Application description
	commands    []*Command                                        // Commands
	stdout      io.Writer                                         // Standard output
	stderr      io.Writer                                         // Standard error
	messageMap  map[errorType]func(app *App, err CLIError) string // Custom message map
}

type Command struct {
	name        string                                        // Command name
	alias       string                                        // Optional command alias
	short       string                                        // Short description shown in global help
	long        string                                        // Detailed description shown in command help
	subcommands []*Command                                    // Nested subcommands
	options     []*Option                                     // Command-specific options
	minArg      int                                           // Minimum required argument
	maxArg      int                                           // Maximum required argument
	action      func(args []string, options map[string]Value) // Execution handler
	parent      *Command                                      // Internal pointer for help generation
}

type Option struct {
	name        string     // Option name
	alias       string     // Optional option alias
	optionType  optionType // Option value type
	value       any        // Default value and type
	description string     // Description shown in help
}

type optionBuilder struct {
	target *Option
	parent *Command
}

type Value struct {
	value any // Store option value
}

// Option types
type optionType int

const (
	String optionType = iota
	Int
	Float
	Bool
)

// Error types
type errorType int

const (
	ErrHelp errorType = iota
	ErrCommandHelp
	ErrVersion
	ErrNoCommand
	ErrUnknownCommand
	ErrSubcommandRequired
	ErrInvalidOption
	ErrOptionValueMissing
	ErrUnexpectedArgument
	ErrTooFewArguments
	ErrTooManyArguments
	ErrInvalidIntValue
	ErrInvalidFloatValue
	ErrInvalidBoolValue
	ErrUnsupportedOptionType
)

type CLIError struct {
	Code    int            // Exit code
	Message string         // Error message
	Type    errorType      // Error type
	Cmd     *Command       // Command pointer
	Data    map[string]any // Extra information
}

func (e CLIError) Error() string {
	return e.Message
}

// The main entry point of the CLI.
func (a *App) Run() int {
	return a.RunWithArgs(os.Args)
}

func (a *App) RunWithArgs(args []string) int {
	if len(args) < 2 {
		return a.stop(ErrNoCommand, nil, nil)
	}

	input := args[1]

	switch input {
	case "--help", "-h":
		return a.stop(ErrHelp, nil, nil)

	case "--version", "-v":
		if a.version != "" {
			return a.stop(ErrVersion, nil, nil)
		}
	}

	// Find the command
	var cmd *Command
	for i := range a.commands {
		if a.commands[i].name == input || a.commands[i].alias == input {
			cmd = a.commands[i]
			break
		}
	}

	if cmd == nil {
		return a.stop(ErrUnknownCommand, nil, map[string]any{
			"cmd": input,
		})
	}

	cmd, remainingArgs := findSubCmd(cmd, args[2:])

	parsedArgs, options, err := a.parseCmd(cmd, remainingArgs)
	if err != -1 {
		return err
	}

	if cmd.action != nil {
		cmd.action(parsedArgs, options)
	}

	return 0
}

// Extracts arguments and options.
func (a *App) parseCmd(cmd *Command, remainingArgs []string) (args []string, options map[string]Value, code int) {
	args = []string{}
	options = make(map[string]Value)

	hasHelpOpt := false
	for _, arg := range remainingArgs {
		if arg == "--" {
			break
		}
		if arg == "--help" || arg == "-h" {
			hasHelpOpt = true
			break
		}
	}

	if hasHelpOpt {
		return nil, nil, a.stop(ErrCommandHelp, cmd, nil)
	}

	// Add default option values to options map
	for _, opt := range cmd.options {
		options[opt.name] = Value{value: opt.value}
	}

	positionalOnly := false

	for i := 0; i < len(remainingArgs); i++ {
		arg := remainingArgs[i]

		if positionalOnly {
			args = append(args, arg)
			continue
		}

		// Enable positional-only mode
		if arg == "--" {
			positionalOnly = true
			continue
		}

		// Positional argument
		if !strings.HasPrefix(arg, "-") {
			args = append(args, arg)
			continue
		}

		// Extract option
		var optName string
		var optValue string
		hasEqualSign := strings.Contains(arg, "=")
		if hasEqualSign {
			parts := strings.SplitN(arg, "=", 2)
			optName = parts[0]
			optValue = parts[1]
		} else {
			optName = arg
		}

		// Option validation
		var matchedOption *Option
		for i := range cmd.options {
			opt := cmd.options[i]
			if optName == "--"+opt.name || (opt.alias != "" && optName == "-"+opt.alias) {
				matchedOption = opt
				break
			}
		}

		if matchedOption == nil {
			return nil, nil, a.stop(ErrInvalidOption, cmd, map[string]any{
				"opt": arg,
			})
		}

		switch matchedOption.optionType {
		case Bool:
			if optValue == "" && !hasEqualSign {
				optValue = "true"
			}
		default:
			if optValue == "" {
				if i+1 < len(remainingArgs) {
					optValue = remainingArgs[i+1]
					i++
				} else {
					return nil, nil, a.stop(ErrOptionValueMissing, cmd, map[string]any{
						"opt": matchedOption.name,
					})
				}
			}
		}

		// Type conversion and validation
		switch matchedOption.optionType {
		case String:
			options[matchedOption.name] = Value{value: optValue}

		case Int:
			parsed, err := strconv.Atoi(optValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidIntValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.name] = Value{value: parsed}

		case Float:
			parsed, err := strconv.ParseFloat(optValue, 64)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidFloatValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.name] = Value{value: parsed}

		case Bool:
			parsed, err := strconv.ParseBool(optValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidBoolValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.name] = Value{value: parsed}

		default:
			return nil, nil, a.stop(ErrUnsupportedOptionType, cmd, map[string]any{
				"val": matchedOption.value,
			})
		}
	}

	// If a command has subcommands but no defined options or run function,
	// then a subcommand is required
	if len(cmd.subcommands) > 0 && len(cmd.options) == 0 && cmd.action == nil {
		return nil, nil, a.stop(ErrSubcommandRequired, cmd, map[string]any{
			"cmd": cmd.name,
		})
	}

	nargs := len(args)

	if cmd.maxArg == 0 && cmd.minArg == 0 && nargs > 0 {
		return nil, nil, a.stop(ErrUnexpectedArgument, cmd, map[string]any{
			"got": nargs,
		})
	}
	if cmd.minArg > 0 && nargs < cmd.minArg {
		return nil, nil, a.stop(ErrTooFewArguments, cmd, map[string]any{
			"got": nargs,
		})
	}
	if cmd.maxArg > 0 && nargs > cmd.maxArg {
		return nil, nil, a.stop(ErrTooManyArguments, cmd, map[string]any{
			"got": nargs,
		})
	}

	return args, options, -1
}

// Recursively resolves nested subcommands.
func findSubCmd(cmd *Command, args []string) (*Command, []string) {
	if len(args) == 0 {
		return cmd, args
	}

	next := args[0]

	for i := range cmd.subcommands {
		sc := cmd.subcommands[i]
		sc.parent = cmd
		if sc.name == next || sc.alias == next {
			return findSubCmd(sc, args[1:])
		}
	}

	return cmd, args
}

// Generates and returns the global help menu.
func (a *App) Help() string {
	var sb strings.Builder

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s [COMMAND] [OPTIONS] [ARGS...]\n", a.name)
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Commands:")

	for _, cmd := range a.commands {
		entry := cmd.name

		if cmd.alias != "" {
			entry += ", " + cmd.alias
		}

		fmt.Fprintf(&sb, "  %-18s %s\n", entry, cmd.short)
	}

	fmt.Fprintln(&sb, "\nOptions:")
	fmt.Fprintf(&sb, "  %-18s %s\n", "--help, -h", "Show help")
	if a.version != "" {
		fmt.Fprintf(&sb, "  %-18s %s\n", "--version, -v", "Show version")
	}

	if len(a.commands) > 0 {
		fmt.Fprintf(&sb, "\nFor more information about a command, use '%s <command> --help'.\n", a.name)
	}

	return sb.String()
}

// Generates and returns a help menu for a specific command.
func (a *App) CommandHelp(cmd *Command) string {
	var sb strings.Builder

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s", a.name)

	// Build full command path
	var parents []string
	currCmd := cmd
	for currCmd.Parent() != nil {
		currCmd = currCmd.Parent()
		parents = append(parents, currCmd.name)
	}

	for i := len(parents) - 1; i >= 0; i-- {
		fmt.Fprintf(&sb, " %s", parents[i])
	}

	fmt.Fprintf(&sb, " %s", cmd.name)

	hasSubcmd := len(cmd.subcommands) > 0
	hasOption := len(cmd.options) > 0

	if hasSubcmd {
		fmt.Fprint(&sb, " [COMMAND]")
	}
	if hasOption {
		fmt.Fprint(&sb, " [OPTIONS]")
	}

	if cmd.maxArg == 1 {
		fmt.Fprint(&sb, " [ARG]")
	} else if cmd.maxArg > 1 {
		fmt.Fprintf(&sb, " [ARG1...ARG%d]", cmd.maxArg)
	} else if cmd.minArg > 0 {
		fmt.Fprint(&sb, " [ARGS...]")
	}

	fmt.Fprintln(&sb)
	if cmd.long != "" {
		fmt.Fprintf(&sb, "\n%s\n", cmd.long)
	}

	if hasSubcmd {
		fmt.Fprintln(&sb, "\nCommands:")
		for _, f := range cmd.subcommands {
			displayName := f.name
			if f.alias != "" {
				displayName += ", " + f.alias
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.short)
		}
	}

	if hasOption {
		fmt.Fprintln(&sb, "\nOptions:")
		for _, f := range cmd.options {
			displayName := "--" + f.name
			if f.alias != "" {
				displayName += ", -" + f.alias
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.description)
		}
	}

	return sb.String()
}

func getMessageAndExitCode(a *App, errType errorType, cmd *Command, data map[string]any) (int, string) {
	var code int = 2
	var msg string

	switch errType {
	case ErrHelp:
		code = 0
		msg = a.Help()

	case ErrCommandHelp:
		code = 0
		msg = a.CommandHelp(cmd)

	case ErrVersion:
		code = 0
		msg = fmt.Sprintf("%s version %s\n", a.name, a.version)

	case ErrNoCommand:
		code = 0
		msg = a.Help()

	case ErrUnknownCommand:
		msg = fmt.Sprintf("unknown command: '%s'\n", data["cmd"])

	case ErrSubcommandRequired:
		msg = fmt.Sprintf("%s requires a subcommand\n", data["cmd"])

	case ErrInvalidOption:
		msg = fmt.Sprintf("invalid option: '%s'\n", data["opt"])

	case ErrOptionValueMissing:
		msg = fmt.Sprintf("value required for option: '%s'\n", data["opt"])

	case ErrUnexpectedArgument:
		msg = fmt.Sprintf("%s does not accept argument(s)\n", cmd.name)

	case ErrTooFewArguments:
		msg = fmt.Sprintf("%s requires at least %d argument(s), got %d\n", cmd.name, cmd.minArg, data["got"])

	case ErrTooManyArguments:
		msg = fmt.Sprintf("%s requires at most %d argument(s), got %d\n", cmd.name, cmd.maxArg, data["got"])

	case ErrInvalidIntValue:
		msg = fmt.Sprintf("int parse error: '%s'\n", data["val"])

	case ErrInvalidFloatValue:
		msg = fmt.Sprintf("float parse error: '%s'\n", data["val"])

	case ErrInvalidBoolValue:
		msg = fmt.Sprintf("bool parse error: '%s'\n", data["val"])

	case ErrUnsupportedOptionType:
		msg = fmt.Sprintf("unsupported type: '%s'\n", fmt.Sprintf("%T", data["val"]))
	default:
		msg = "unknown error"
	}

	return code, msg
}

func (a *App) stop(errType errorType, cmd *Command, data map[string]any) int {
	code, message := getMessageAndExitCode(a, errType, cmd, data)

	return a.handleError(CLIError{
		Code:    code,
		Type:    errType,
		Cmd:     cmd,
		Message: message,
		Data:    data,
	})
}

func (a *App) handleError(err CLIError) int {
	var msg string
	out := a.Stderr()

	if a.messageMap != nil {
		if fn, ok := a.messageMap[err.Type]; ok {
			msg = fn(a, err)
		}
	}

	if msg == "" {
		msg = err.Message
	}

	if err.Code == 0 {
		out = a.Stdout()
	}

	fmt.Fprint(out, msg)
	return err.Code
}

func (a *App) HandleMessage(errType errorType, fn func(app *App, err CLIError) string) *App {
	a.messageMap[errType] = fn
	return a
}

// Returns string value.
func (v Value) String() string {
	s := v.value.(string)
	return s
}

// Returns int value.
func (v Value) Int() int {
	i := v.value.(int)
	return i
}

// Returns float value.
func (v Value) Float() float64 {
	f := v.value.(float64)
	return f
}

// Returns bool value.
func (v Value) Bool() bool {
	b := v.value.(bool)
	return b
}

func NewApp(name string) *App {
	return &App{
		name:       name,
		commands:   []*Command{},
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		messageMap: map[errorType]func(app *App, err CLIError) string{},
	}
}

func (a *App) SetVersion(version string) *App {
	a.version = version
	return a
}

func (a *App) SetDescription(description string) *App {
	a.description = description
	return a
}

func (a *App) SetStdout(out io.Writer) *App {
	a.stdout = out
	return a
}

func (a *App) SetStderr(err io.Writer) *App {
	a.stderr = err
	return a
}

func (a *App) AddCommand(name string) *Command {
	c := &Command{
		name:        name,
		options:     []*Option{},
		subcommands: []*Command{},
	}
	a.commands = append(a.commands, c)
	return c
}

func (c *Command) SetAlias(alias string) *Command {
	c.alias = alias
	return c
}

func (c *Command) SetShort(short string) *Command {
	c.short = short
	return c
}

func (c *Command) SetLong(long string) *Command {
	c.long = long
	return c
}

func (c *Command) SetMinArg(min int) *Command {
	c.minArg = min
	return c
}

func (c *Command) SetMaxArg(max int) *Command {
	c.maxArg = max
	return c
}

func (c *Command) AddSubcommand(name string) *Command {
	sub := &Command{name: name, parent: c}
	c.subcommands = append(c.subcommands, sub)
	return sub
}

func (c *Command) Ok() *Command {
	return c.parent
}

func (c *Command) AddOption(name string) *optionBuilder {
	o := &Option{name: name}
	return &optionBuilder{
		target: o,
		parent: c,
	}
}

func (o *optionBuilder) SetAlias(alias string) *optionBuilder {
	o.target.alias = alias
	return o
}

func (o *optionBuilder) SetType(optionType optionType) *optionBuilder {
	o.target.optionType = optionType
	return o
}

func (o *optionBuilder) SetValue(value any) *optionBuilder {
	o.target.value = value
	return o
}

func (o *optionBuilder) SetDescription(description string) *optionBuilder {
	o.target.description = description
	return o
}

func (o *optionBuilder) Ok() *Command {
	o.parent.options = append(o.parent.options, o.target)
	return o.parent
}

func (c *Command) Action(fn func(args []string, options map[string]Value)) *Command {
	c.action = fn
	return c
}

func (a *App) Name() string         { return a.name }
func (a *App) Version() string      { return a.version }
func (a *App) Description() string  { return a.description }
func (a *App) Commands() []*Command { return a.commands }

func (a *App) Stdout() io.Writer {
	if a.stdout != nil {
		return a.stdout
	}
	return os.Stdout
}
func (a *App) Stderr() io.Writer {
	if a.stderr != nil {
		return a.stderr
	}
	return os.Stderr
}

func (c *Command) Name() string            { return c.name }
func (c *Command) Alias() string           { return c.alias }
func (c *Command) Short() string           { return c.short }
func (c *Command) Long() string            { return c.long }
func (c *Command) Subcommands() []*Command { return c.subcommands }
func (c *Command) Options() []*Option      { return c.options }
func (c *Command) MinArg() int             { return c.minArg }
func (c *Command) MaxArg() int             { return c.maxArg }
func (c *Command) Parent() *Command        { return c.parent }

func (o *Option) Name() string           { return o.name }
func (o *Option) Alias() string          { return o.alias }
func (o *Option) OptionType() optionType { return o.optionType }
func (o *Option) Value() any             { return o.value }
func (o *Option) Description() string    { return o.description }
