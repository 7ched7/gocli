package gocli

import (
	"flag"
	"fmt"
	"io"
	"os"
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
	Name  string // Option name
	Alias string // Optional option alias
	Value any    // Default value and type
	Desc  string // Description shown in help
}

type Value struct {
	Any any // Store option value
}

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

	cmd, remainingArgs := findCmd(cmd, args[2:])

	parsedArgs, options, err := a.parseCmd(cmd, remainingArgs)
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
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

// Recursively resolves nested subcommands.
func findCmd(cmd *Command, args []string) (*Command, []string) {
	if len(args) == 0 {
		return cmd, args
	}

	next := args[0]

	for i := range cmd.Subcommand {
		sc := &cmd.Subcommand[i]
		sc.parent = cmd
		if sc.Name == next || sc.Alias == next {
			return findCmd(sc, args[1:])
		}
	}

	return cmd, args
}

// Registers and parses options for a command.
func (a *App) parseCmd(cmd *Command, args []string) ([]string, map[string]Value, error) {
	flagSet := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	flagSet.SetOutput(a.stderr())

	flagSet.Usage = func() {
		a.printCmdHelp(cmd, flagSet.Output())
	}

	values := make(map[string]any)

	// Create options
	for _, f := range cmd.Options {
		switch v := f.Value.(type) {
		case bool:
			p := flagSet.Bool(f.Name, v, f.Desc)
			if f.Alias != "" {
				flagSet.BoolVar(p, f.Alias, v, f.Desc)
			}
			values[f.Name] = p
		case int:
			p := flagSet.Int(f.Name, v, f.Desc)
			if f.Alias != "" {
				flagSet.IntVar(p, f.Alias, v, f.Desc)
			}
			values[f.Name] = p
		case string:
			p := flagSet.String(f.Name, v, f.Desc)
			if f.Alias != "" {
				flagSet.StringVar(p, f.Alias, v, f.Desc)
			}
			values[f.Name] = p
		}
	}

	err := flagSet.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	// Convert option pointers into a key/value map
	// according to their types
	options := make(map[string]Value)
	for key, val := range values {
		switch v := val.(type) {
		case *bool:
			options[key] = Value{Any: *v}
		case *int:
			options[key] = Value{Any: *v}
		case *string:
			options[key] = Value{Any: *v}
		}
	}

	return flagSet.Args(), options, nil
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
func (v Value) String() string {
	s := v.Any.(string)
	return s
}

// Returns int value.
func (v Value) Int() int {
	i := v.Any.(int)
	return i
}

// Returns bool value.
func (v Value) Bool() bool {
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
