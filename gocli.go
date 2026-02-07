package gocli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type App struct {
	Name     string    // Application name
	Version  string    // Optional app version
	Commands []Command // Commands
	Stdout   io.Writer // Standard output
	Stderr   io.Writer // Standard error
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
	Bool
)

// The main entry point of the CLI.
func (a *App) Run() int {
	return a.RunWithArgs(os.Args)
}

func (a *App) RunWithArgs(args []string) int {
	if len(args) < 2 {
		a.printHelp(a.stdout())
		return 0
	}

	input := args[1]

	switch input {
	case "--help", "-h":
		a.printHelp(a.stdout())
		return 0
	case "--version", "-v":
		if a.Version != "" {
			fmt.Fprintf(a.stdout(), "%s version %s\n", a.Name, a.Version)
			return 0
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
		fmt.Fprintf(a.stderr(), "unknown command: %s\n", input)
		a.printHelp(a.stderr())
		return 2
	}

	cmd, remainingArgs := findSubCmd(cmd, args[2:])

	parsedArgs, options, err := a.parseCmd(cmd, remainingArgs)
	if err != nil {
		fmt.Fprint(a.stderr(), err)
		return 2
	}

	// If a command has subcommands but no defined options or run function,
	// then a subcommand is required
	if len(cmd.Subcommand) > 0 && len(cmd.Options) == 0 && cmd.Run == nil {
		fmt.Fprintf(a.stderr(), "%s requires a subcommand\n", cmd.Name)
		a.printCmdHelp(cmd, a.stderr())
		return 2
	}

	nargs := len(parsedArgs)

	if cmd.MaxArg == 0 && cmd.MinArg == 0 && nargs > 0 {
		fmt.Fprintf(a.stderr(), "%s does not accept argument(s)\n", cmd.Name)
		a.printCmdHelp(cmd, a.stderr())
		return 2
	}
	if cmd.MinArg > 0 && nargs < cmd.MinArg {
		fmt.Fprintf(a.stderr(), "%s requires at least %d argument(s), got %d\n", cmd.Name, cmd.MinArg, nargs)
		a.printCmdHelp(cmd, a.stderr())
		return 2
	}
	if cmd.MaxArg > 0 && nargs > cmd.MaxArg {
		fmt.Fprintf(a.stderr(), "%s requires at most %d argument(s), got %d\n", cmd.Name, cmd.MaxArg, nargs)
		a.printCmdHelp(cmd, a.stderr())
		return 2
	}

	if cmd.Run != nil {
		cmd.Run(parsedArgs, options)
	}

	return 0
}

// Extracts arguments and options.
func (a *App) parseCmd(cmd *Command, remainingArgs []string) (args []string, options map[string]Value, err error) {
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
		a.printCmdHelp(cmd, a.stdout())
		return nil, nil, fmt.Errorf("")
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
			return nil, nil, fmt.Errorf("invalid option: '%s'\n", arg)
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
					return nil, nil, fmt.Errorf("value required for option: '%s'\n", matchedOption.Name)
				}
			}
		}

		// Type conversion and validation
		switch matchedOption.Type {
		case Bool:
			parsed, err := strconv.ParseBool(optValue)
			if err != nil {
				return nil, nil, fmt.Errorf("bool parse error: %s\n", optValue)
			}
			options[matchedOption.Name] = Value{Any: parsed}

		case Int:
			parsed, err := strconv.Atoi(optValue)
			if err != nil {
				return nil, nil, fmt.Errorf("int parse error: %s\n", optValue)
			}
			options[matchedOption.Name] = Value{Any: parsed}

		case String:
			options[matchedOption.Name] = Value{Any: optValue}

		default:
			return nil, nil, fmt.Errorf("unsupported type: %s\n", fmt.Sprintf("%T", matchedOption.Value))
		}
	}

	return args, options, nil
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

// Displays the global help menu.
func (a *App) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  %s [COMMAND] [OPTIONS] [ARGS...]\n", a.Name)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")

	for _, cmd := range a.Commands {
		entry := cmd.Name

		if cmd.Alias != "" {
			entry += ", " + cmd.Alias
		}

		fmt.Fprintf(w, "  %-18s %s\n", entry, cmd.Short)
	}

	fmt.Fprintln(w, "\nOptions:")
	fmt.Fprintf(w, "  %-18s %s\n", "--help, -h", "Show help")
	if a.Version != "" {
		fmt.Fprintf(w, "  %-18s %s\n", "--version, -v", "Show version")
	}

	if len(a.Commands) > 0 {
		fmt.Fprintf(w, "\nFor more information about a command, use '%s <command> --help'.\n", a.Name)
	}
}

// Displays the command-specific help menu.
func (a *App) printCmdHelp(cmd *Command, w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintf(w, "  %s", a.Name)

	// Build full command path
	var parents []string
	currCmd := cmd
	for currCmd.parent != nil {
		currCmd = currCmd.parent
		parents = append(parents, currCmd.Name)
	}

	for i := len(parents) - 1; i >= 0; i-- {
		fmt.Fprintf(w, " %s", parents[i])
	}

	fmt.Fprintf(w, " %s", cmd.Name)

	hasSubcmd := len(cmd.Subcommand) > 0
	hasOption := len(cmd.Options) > 0

	if hasSubcmd {
		fmt.Fprint(w, " [COMMAND]")
	}
	if hasOption {
		fmt.Fprint(w, " [OPTIONS]")
	}

	if cmd.MaxArg == 1 {
		fmt.Fprint(w, " [ARG]")
	} else if cmd.MaxArg > 1 {
		fmt.Fprintf(w, " [ARG1...ARG%d]", cmd.MaxArg)
	} else if cmd.MinArg > 0 {
		fmt.Fprint(w, " [ARGS...]")
	}

	fmt.Fprintln(w)
	if cmd.Long != "" {
		fmt.Fprintf(w, "\n%s\n", cmd.Long)
	}

	if hasSubcmd {
		fmt.Fprintln(w, "\nCommands:")
		for _, f := range cmd.Subcommand {
			displayName := f.Name
			if f.Alias != "" {
				displayName += ", " + f.Alias
			}
			fmt.Fprintf(w, "  %-18s %s\n", displayName, f.Short)
		}
	}

	if hasOption {
		fmt.Fprintln(w, "\nOptions:")
		for _, f := range cmd.Options {
			displayName := "--" + f.Name
			if f.Alias != "" {
				displayName += ", -" + f.Alias
			}
			fmt.Fprintf(w, "  %-18s %s\n", displayName, f.Desc)
		}
	}
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

// Returns bool value.
func (v Value) GetBool() bool {
	b := v.Any.(bool)
	return b
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
