package gocli

import (
	"strings"
)

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

	// If a command has subcommands but no defined flags or action function,
	// then a subcommand is required
	if len(cmd.subcommands) > 0 && len(cmd.flags) == 0 && cmd.action == nil {
		return nil, nil, a.stop(ErrSubcommandRequired, cmd, map[string]any{
			"command": cmd.name,
		})
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

		var newi int
		var code int

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, code = a.handleShortFlag(cmd, flags, arg, remainingArgs, i)
		} else {
			newi, code = a.handleLongFlag(cmd, flags, arg, remainingArgs, i)
		}

		if code != -1 {
			return nil, nil, code
		}

		i = newi
	}

	if code := a.validateArgCount(cmd, len(args)); code != -1 {
		return nil, nil, code
	}

	return args, flags, -1
}

func (a *App) findFlagName(cmd *Command, flagName string) (*Flag, int) {
	var matchedFlag *Flag
	for i := range cmd.flags {
		flag := cmd.flags[i]
		if flagName == "--"+flag.name || (flag.alias != "" && flagName == "-"+flag.alias) {
			matchedFlag = flag
			break
		}
	}

	if matchedFlag == nil {
		return nil, a.stop(ErrInvalidFlag, cmd, map[string]any{
			"flag": flagName,
		})
	}

	return matchedFlag, -1
}

func (a *App) handleLongFlag(cmd *Command, flags *Flags, arg string, remainingArgs []string, i int) (int, int) {
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

	matchedFlag, code := a.findFlagName(cmd, flagName)
	if code != -1 {
		return i, code
	}

	switch matchedFlag.flagType {
	case Bool:
		if flagValue == "" && !hasEqualSign {
			flagValue = "true"
		}
	default:
		if flagValue == "" {
			if i+1 < len(remainingArgs) && !hasEqualSign { // --flag value
				flagValue = remainingArgs[i+1]
				i++
			} else {
				return i, a.stop(ErrFlagValueMissing, cmd, map[string]any{
					"flag": matchedFlag.name,
				})
			}
		}
	}

	if code := a.setFlagValue(cmd, matchedFlag, flags, flagValue); code != -1 {
		return i, code
	}

	return i, -1
}

func (a *App) handleShortFlag(cmd *Command, flags *Flags, arg string, remainingArgs []string, i int) (int, int) {
	for j, f := range arg[1:] {
		matchedFlag, code := a.findFlagName(cmd, "-"+string(f))
		if code != -1 {
			return i, code
		}

		var flagValue string

		switch matchedFlag.flagType {
		case Bool:
			flagValue = "true"
		default:
			if j < len(arg[1:])-1 { // -fvalue
				flagValue = arg[j+2:]
			} else if i+1 < len(remainingArgs) { // -f value
				flagValue = remainingArgs[i+1]
				i++
			} else {
				return i, a.stop(ErrFlagValueMissing, cmd, map[string]any{
					"flag": matchedFlag.alias,
				})
			}
		}

		if code := a.setFlagValue(cmd, matchedFlag, flags, flagValue); code != -1 {
			return i, code
		}

		if matchedFlag.flagType != Bool {
			break
		}
	}

	return i, -1
}

func (a *App) setFlagValue(cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
	switch matchedFlag.flagType {
	case StringSlice, IntSlice, FloatSlice, BoolSlice:
		for _, v := range strings.Split(flagValue, ",") {
			if code := a.setFlagValueByType(cmd, matchedFlag, flags, v); code != -1 {
				return code
			}
		}
	default:
		if code := a.setFlagValueByType(cmd, matchedFlag, flags, flagValue); code != -1 {
			return code
		}
	}
	return -1
}

func (a *App) setFlagValueByType(cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
	fn, ok := flagTypeHandlerMap[matchedFlag.flagType]
	if !ok {
		return a.stop(ErrUnsupportedFlagType, cmd, map[string]any{
			"value": matchedFlag.flagType,
		})
	}

	if code := fn(a, cmd, matchedFlag, flags, flagValue); code != -1 {
		return code
	}

	return -1
}

func (a *App) validateArgCount(cmd *Command, nargs int) int {
	if cmd.maxArg == 0 && cmd.minArg == 0 && nargs > 0 {
		return a.stop(ErrUnexpectedArgument, cmd, map[string]any{
			"number": nargs,
		})
	}
	if cmd.minArg > 0 && nargs < cmd.minArg {
		return a.stop(ErrTooFewArguments, cmd, map[string]any{
			"number": nargs,
		})
	}
	if cmd.maxArg > 0 && nargs > cmd.maxArg {
		return a.stop(ErrTooManyArguments, cmd, map[string]any{
			"number": nargs,
		})
	}

	return -1
}

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

func (a *App) findRootCommand(input string) (*Command, int) {
	for i := range a.commands {
		if a.commands[i].name == input || a.commands[i].alias == input {
			return a.commands[i], -1
		}
	}
	return nil, a.stop(ErrUnknownCommand, nil, map[string]any{
		"command": input,
	})
}

func (a *App) findSubcommand(cmd *Command, args []string) (*Command, []string) {
	if len(args) == 0 {
		return cmd, args
	}

	next := args[0]

	for i := range cmd.subcommands {
		sc := cmd.subcommands[i]
		if sc.name == next || sc.alias == next {
			return a.findSubcommand(sc, args[1:])
		}
	}

	return cmd, args
}

func (a *App) runCommand(
	cmd *Command,
	args []string,
	flags Flags,
) {
	if cmd.action != nil {
		cmd.action(args, flags)
	}
}
