package gocli

import (
	"strconv"
	"strings"
)

// parseCommand processes commands entered by the user,
// handles different incorrect command input scenarios, edge cases,
// and finally returns positional arguments and parsed flags.
func (a *App) parseCommand(cmd *Command, remainingArgs []string) ([]string, *Flags, int) {
	args := []string{}
	flags := &Flags{pair: map[string]any{}}

	hasHelpFlag := false
	for _, arg := range remainingArgs {
		if arg == "--" {
			break
		}
		if arg == "--help" || arg == "-h" {
			hasHelpFlag = true
			break
		}
	}

	if hasHelpFlag {
		return nil, nil, a.stop(ErrCommandHelp, cmd, nil)
	}

	// Add default flag values to flags map
	for _, flag := range cmd.flags {
		flags.pair[flag.name] = flag.defaultValue
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

		// Extract flag
		var flagName string
		var flagValue string
		hasEqualSign := strings.Contains(arg, "=")
		if hasEqualSign {
			parts := strings.SplitN(arg, "=", 2)
			flagName = parts[0]
			flagValue = parts[1]
		} else {
			flagName = arg
		}

		// Flag validation
		var matchedFlag *Flag
		for i := range cmd.flags {
			flag := cmd.flags[i]
			if flagName == "--"+flag.name || (flag.alias != "" && flagName == "-"+flag.alias) {
				matchedFlag = flag
				break
			}
		}

		if matchedFlag == nil {
			return nil, nil, a.stop(ErrInvalidFlag, cmd, map[string]any{
				"flag": arg,
			})
		}

		switch matchedFlag.flagType {
		case Bool:
			if flagValue == "" && !hasEqualSign {
				flagValue = "true"
			}
		default:
			if flagValue == "" {
				if i+1 < len(remainingArgs) {
					flagValue = remainingArgs[i+1]
					i++
				} else {
					return nil, nil, a.stop(ErrFlagValueMissing, cmd, map[string]any{
						"flag": matchedFlag.name,
					})
				}
			}
		}

		// Type conversion and validation
		switch matchedFlag.flagType {
		case String:
			flags.pair[matchedFlag.name] = flagValue

		case Int:
			parsed, err := strconv.Atoi(flagValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidIntValue, cmd, map[string]any{
					"val": flagValue,
				})
			}

			flags.pair[matchedFlag.name] = parsed

		case Float:
			parsed, err := strconv.ParseFloat(flagValue, 64)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidFloatValue, cmd, map[string]any{
					"val": flagValue,
				})
			}

			flags.pair[matchedFlag.name] = parsed

		case Bool:
			parsed, err := strconv.ParseBool(flagValue)
			if err != nil {
				return nil, nil, a.stop(ErrInvalidBoolValue, cmd, map[string]any{
					"val": flagValue,
				})
			}

			flags.pair[matchedFlag.name] = parsed

		default:
			return nil, nil, a.stop(ErrUnsupportedFlagType, cmd, map[string]any{
				"val": matchedFlag.defaultValue,
			})
		}
	}

	// If a command has subcommands but no defined flags or action function,
	// then a subcommand is required
	if len(cmd.subcommands) > 0 && len(cmd.flags) == 0 && cmd.action == nil {
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

	return args, flags, -1
}

// handleGlobalArgs handles app-specific --help and --version flags.
func (a *App) handleGlobalArgs(args []string) int {
	if len(args) < 2 {
		return a.stop(ErrNoCommand, nil, nil)
	}

	switch args[1] {
	case "--help", "-h":
		return a.stop(ErrHelp, nil, nil)

	case "--version", "-v":
		if a.version != "" {
			return a.stop(ErrVersion, nil, nil)
		}
	}

	return -1
}

// findRootCommand finds and returns the top-level command.
func (a *App) findRootCommand(input string) (*Command, int) {
	for i := range a.commands {
		if a.commands[i].name == input || a.commands[i].alias == input {
			return a.commands[i], -1
		}
	}
	return nil, a.stop(ErrUnknownCommand, nil, map[string]any{
		"cmd": input,
	})
}

// findSubcommand recursively resolves nested subcommands.
func (a *App) findSubcommand(cmd *Command, args []string) (*Command, []string) {
	if len(args) == 0 {
		return cmd, args
	}

	next := args[0]

	for i := range cmd.subcommands {
		sc := cmd.subcommands[i]
		sc.parent = cmd
		if sc.name == next || sc.alias == next {
			return a.findSubcommand(sc, args[1:])
		}
	}

	return cmd, args
}

// runCommand executes the action function of a command,
// with the positional arguments and parsed args.
func (a *App) runCommand(
	cmd *Command,
	args []string,
	flags Flags,
) {
	if cmd.action != nil {
		cmd.action(args, flags)
	}
}
