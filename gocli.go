package gocli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type App struct {
	Name           string                                            // Application name
	Version        string                                            // Optional app version
	Commands       []Command                                         // Commands
	Stdout         io.Writer                                         // Standard output
	Stderr         io.Writer                                         // Standard error
	MessageHandler map[ErrorType]func(app *App, err CLIError) string // Custom message handler
}

type Command struct {
	Name       string                                        // Command name
	Alias      string                                        // Optional command alias
	Short      string                                        // Short description shown in global help
	Long       string                                        // Detailed description shown in command help
	Subcommand []Command                                     // Nested subcommands
	Options    []Option                                      // Command-specific options
	MinArg     int                                           // Minimum required argument
	MaxArg     int                                           // Maximum required argument
	Run        func(args []string, options map[string]Value) // Execution handler
	parent     *Command                                      // Internal pointer for help generation
}

type Option struct {
	Name  string     // Option name
	Alias string     // Optional option alias
	Type  OptionType // Option type
	Value any        // Default value and type
	Desc  string     // Description shown in help
}

type Value struct {
	Any any // Store option value
}

// Option types
type OptionType int

const (
	String OptionType = iota
	Int
	Float
	Bool
)

// Error types
type ErrorType int

const (
	ErrHelp ErrorType = iota
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
	Type    ErrorType      // Error type
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
		if a.Version != "" {
			return a.stop(ErrVersion, nil, nil)
		}
	}

	// Find the command
	var cmd *Command
	for i := range a.Commands {
		if a.Commands[i].Name == input || a.Commands[i].Alias == input {
			cmd = &a.Commands[i]
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

	if cmd.Run != nil {
		cmd.Run(parsedArgs, options)
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
	for _, opt := range cmd.Options {
		options[opt.Name] = Value{Any: opt.Value}
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
		for i := range cmd.Options {
			opt := &cmd.Options[i]
			if optName == "--"+opt.Name || (opt.Alias != "" && optName == "-"+opt.Alias) {
				matchedOption = opt
				break
			}
		}

		if matchedOption == nil {
			return nil, nil, a.stop(ErrInvalidOption, cmd, map[string]any{
				"opt": arg,
			})
		}

		switch matchedOption.Type {
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
						"opt": matchedOption.Name,
					})
				}
			}
		}

		// Type conversion and validation
		switch matchedOption.Type {
		case String:
			options[matchedOption.Name] = Value{Any: optValue}

		case Int:
			parsed, err := strconv.Atoi(optValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidIntValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.Name] = Value{Any: parsed}

		case Float:
			parsed, err := strconv.ParseFloat(optValue, 64)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidFloatValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.Name] = Value{Any: parsed}

		case Bool:
			parsed, err := strconv.ParseBool(optValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidBoolValue, cmd, map[string]any{
					"val": optValue,
				})
			}

			options[matchedOption.Name] = Value{Any: parsed}

		default:
			return nil, nil, a.stop(ErrUnsupportedOptionType, cmd, map[string]any{
				"val": matchedOption.Value,
			})
		}
	}

	// If a command has subcommands but no defined options or run function,
	// then a subcommand is required
	if len(cmd.Subcommand) > 0 && len(cmd.Options) == 0 && cmd.Run == nil {
		return nil, nil, a.stop(ErrSubcommandRequired, cmd, map[string]any{
			"cmd": cmd.Name,
		})
	}

	nargs := len(args)

	if cmd.MaxArg == 0 && cmd.MinArg == 0 && nargs > 0 {
		return nil, nil, a.stop(ErrUnexpectedArgument, cmd, map[string]any{
			"got": nargs,
		})
	}
	if cmd.MinArg > 0 && nargs < cmd.MinArg {
		return nil, nil, a.stop(ErrTooFewArguments, cmd, map[string]any{
			"got": nargs,
		})
	}
	if cmd.MaxArg > 0 && nargs > cmd.MaxArg {
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

	for i := range cmd.Subcommand {
		sc := &cmd.Subcommand[i]
		sc.parent = cmd
		if sc.Name == next || sc.Alias == next {
			return findSubCmd(sc, args[1:])
		}
	}

	return cmd, args
}

// Generates and returns the global help menu.
func (a *App) Help() string {
	var sb strings.Builder

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s [COMMAND] [OPTIONS] [ARGS...]\n", a.Name)
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Commands:")

	for _, cmd := range a.Commands {
		entry := cmd.Name

		if cmd.Alias != "" {
			entry += ", " + cmd.Alias
		}

		fmt.Fprintf(&sb, "  %-18s %s\n", entry, cmd.Short)
	}

	fmt.Fprintln(&sb, "\nOptions:")
	fmt.Fprintf(&sb, "  %-18s %s\n", "--help, -h", "Show help")
	if a.Version != "" {
		fmt.Fprintf(&sb, "  %-18s %s\n", "--version, -v", "Show version")
	}

	if len(a.Commands) > 0 {
		fmt.Fprintf(&sb, "\nFor more information about a command, use '%s <command> --help'.\n", a.Name)
	}

	return sb.String()
}

// Generates and returns a help menu for a specific command.
func (a *App) CommandHelp(cmd *Command) string {
	var sb strings.Builder

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s", a.Name)

	// Build full command path
	var parents []string
	currCmd := cmd
	for currCmd.Parent() != nil {
		currCmd = currCmd.Parent()
		parents = append(parents, currCmd.Name)
	}

	for i := len(parents) - 1; i >= 0; i-- {
		fmt.Fprintf(&sb, " %s", parents[i])
	}

	fmt.Fprintf(&sb, " %s", cmd.Name)

	hasSubcmd := len(cmd.Subcommand) > 0
	hasOption := len(cmd.Options) > 0

	if hasSubcmd {
		fmt.Fprint(&sb, " [COMMAND]")
	}
	if hasOption {
		fmt.Fprint(&sb, " [OPTIONS]")
	}

	if cmd.MaxArg == 1 {
		fmt.Fprint(&sb, " [ARG]")
	} else if cmd.MaxArg > 1 {
		fmt.Fprintf(&sb, " [ARG1...ARG%d]", cmd.MaxArg)
	} else if cmd.MinArg > 0 {
		fmt.Fprint(&sb, " [ARGS...]")
	}

	fmt.Fprintln(&sb)
	if cmd.Long != "" {
		fmt.Fprintf(&sb, "\n%s\n", cmd.Long)
	}

	if hasSubcmd {
		fmt.Fprintln(&sb, "\nCommands:")
		for _, f := range cmd.Subcommand {
			displayName := f.Name
			if f.Alias != "" {
				displayName += ", " + f.Alias
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.Short)
		}
	}

	if hasOption {
		fmt.Fprintln(&sb, "\nOptions:")
		for _, f := range cmd.Options {
			displayName := "--" + f.Name
			if f.Alias != "" {
				displayName += ", -" + f.Alias
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.Desc)
		}
	}

	return sb.String()
}

func getMessageAndExitCode(a *App, errType ErrorType, cmd *Command, data map[string]any) (int, string) {
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
		msg = fmt.Sprintf("%s version %s\n", a.Name, a.Version)

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
		msg = fmt.Sprintf("%s does not accept argument(s)\n", cmd.Name)

	case ErrTooFewArguments:
		msg = fmt.Sprintf("%s requires at least %d argument(s), got %d\n", cmd.Name, cmd.MinArg, data["got"])

	case ErrTooManyArguments:
		msg = fmt.Sprintf("%s requires at most %d argument(s), got %d\n", cmd.Name, cmd.MaxArg, data["got"])

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

func (a *App) stop(errType ErrorType, cmd *Command, data map[string]any) int {
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
	out := a.stderr()

	if a.MessageHandler != nil {
		if fn, ok := a.MessageHandler[err.Type]; ok {
			msg = fn(a, err)
		}
	}

	if msg == "" {
		msg = err.Message
	}

	if err.Code == 0 {
		out = a.stdout()
	}

	fmt.Fprint(out, msg)
	return err.Code
}

// Returns string value.
func (v Value) GetString() string {
	s := v.Any.(string)
	return s
}

// Returns int value.
func (v Value) GetInt() int {
	i := v.Any.(int)
	return i
}

// Returns float value.
func (v Value) GetFloat() float64 {
	f := v.Any.(float64)
	return f
}

// Returns bool value.
func (v Value) GetBool() bool {
	b := v.Any.(bool)
	return b
}

func (c *Command) Parent() *Command {
	return c.parent
}

func (a *App) stdout() io.Writer {
	if a.Stdout != nil {
		return a.Stdout
	}
	return os.Stdout
}

func (a *App) stderr() io.Writer {
	if a.Stderr != nil {
		return a.Stderr
	}
	return os.Stderr
}
