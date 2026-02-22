package gocli

import (
	"fmt"
	"strings"
)

func (a *App) parseCommand(args []string) (*Command, []string, *Flags, int) {
	if a.root.action == nil && len(args) == 0 {
		return nil, nil, nil, a.stop(ErrNoCommand, nil, nil)
	}

	cmd := a.root
	pargs := []string{}
	flags := &Flags{pair: map[string]any{}}
	positionalOnly := false

	// Default global flag values mapping
	for _, flag := range a.globalFlags {
		flags.pair[flag.name] = flag.defaultValue
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if positionalOnly {
			pargs = append(pargs, arg)
			continue
		}

		// Enable positional-only mode
		if arg == "--" {
			positionalOnly = true
			continue
		}

		var code int
		var newi int

		// Positional argument
		if !strings.HasPrefix(arg, "-") {
			cmd, pargs, code = a.handleArgument(cmd, flags, arg, pargs)
			if code != StateContinue {
				return nil, nil, nil, code
			}

			continue
		}

		code = a.handleHelpAndVersion(arg, cmd)
		if code != StateContinue {
			return nil, nil, nil, code
		}

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, code = a.handleShortFlag(cmd, flags, arg, args, i)
		} else {
			newi, code = a.handleLongFlag(cmd, flags, arg, args, i)
		}

		if code != StateContinue {
			return nil, nil, nil, code
		}

		i = newi
	}

	if len(cmd.subcommands) > 0 && cmd.minArg == 0 && cmd.maxArg == 0 {
		return nil, nil, nil, a.stop(ErrSubcommandRequired, cmd, map[string]string{
			"command": cmd.name,
		})
	}

	if code := a.validateArgCount(cmd, len(pargs)); code != StateContinue {
		return nil, nil, nil, code
	}

	return cmd, pargs, flags, StateContinue
}

func (a *App) handleArgument(cmd *Command, flags *Flags, arg string, pargs []string) (*Command, []string, int) {
	isCmd := false

	if len(pargs) == 0 {
		if cmd == a.root {
			for i := range a.commands {
				if a.commands[i].name == arg || a.commands[i].alias == arg {
					isCmd = true
					cmd = a.commands[i]
				}
			}
		} else {
			for i := range cmd.subcommands {
				if cmd.subcommands[i].name == arg || cmd.subcommands[i].alias == arg {
					isCmd = true
					cmd = cmd.subcommands[i]
				}
			}
		}
	}

	if !isCmd {
		if cmd == a.root && cmd.minArg == 0 && cmd.maxArg == 0 {
			return nil, nil, a.stop(ErrUnknownCommand, cmd, map[string]string{
				"command": arg,
			})
		}

		pargs = append(pargs, arg)
	} else {
		// Default command flag values mapping
		for _, flag := range cmd.flags {
			flags.pair[flag.name] = flag.defaultValue
		}
	}

	return cmd, pargs, StateContinue
}

func (a *App) findFlagName(cmd *Command, flagName string) (*Flag, int) {
	var matchedFlag *Flag

	if cmd != a.root {
		for i := range cmd.flags {
			flag := cmd.flags[i]
			if flagName == "--"+flag.name || (flag.alias != "" && flagName == "-"+flag.alias) {
				matchedFlag = flag
				break
			}
		}
	}

	if matchedFlag == nil {
		for i := range a.globalFlags {
			flag := a.globalFlags[i]
			if flagName == "--"+flag.name || (flag.alias != "" && flagName == "-"+flag.alias) {
				matchedFlag = flag
				break
			}
		}
	}

	if matchedFlag == nil {
		return nil, a.stop(ErrInvalidFlag, cmd, map[string]string{
			"flag": flagName,
		})
	}

	return matchedFlag, StateContinue
}

func (a *App) handleLongFlag(cmd *Command, flags *Flags, arg string, args []string, i int) (int, int) {
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
	if code != StateContinue {
		return i, code
	}

	switch matchedFlag.flagType {
	case Bool:
		if flagValue == "" && !hasEqualSign {
			flagValue = "true"
		}
	default:
		if flagValue == "" {
			if i+1 < len(args) && !hasEqualSign { // --flag value
				flagValue = args[i+1]
				i++
			} else {
				return i, a.stop(ErrFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.name,
				})
			}
		}
	}

	if code := a.setFlagValue(cmd, matchedFlag, flags, flagValue); code != StateContinue {
		return i, code
	}

	return i, StateContinue
}

func (a *App) handleShortFlag(cmd *Command, flags *Flags, arg string, args []string, i int) (int, int) {
	for j, f := range arg[1:] {
		matchedFlag, code := a.findFlagName(cmd, "-"+string(f))
		if code != StateContinue {
			return i, code
		}

		var flagValue string

		switch matchedFlag.flagType {
		case Bool:
			flagValue = "true"
		default:
			if j < len(arg[1:])-1 { // -fvalue
				flagValue = arg[j+2:]
			} else if i+1 < len(args) { // -f value
				flagValue = args[i+1]
				i++
			} else {
				return i, a.stop(ErrFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.alias,
				})
			}
		}

		if code := a.setFlagValue(cmd, matchedFlag, flags, flagValue); code != StateContinue {
			return i, code
		}

		if matchedFlag.flagType != Bool {
			break
		}
	}

	return i, StateContinue
}

func (a *App) setFlagValue(cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
	switch matchedFlag.flagType {
	case StringSlice, IntSlice, FloatSlice, BoolSlice:
		for _, v := range strings.Split(flagValue, ",") {
			if code := a.setFlagValueByType(cmd, matchedFlag, flags, v); code != StateContinue {
				return code
			}
		}
	default:
		if code := a.setFlagValueByType(cmd, matchedFlag, flags, flagValue); code != StateContinue {
			return code
		}
	}
	return StateContinue
}

func (a *App) setFlagValueByType(cmd *Command, matchedFlag *Flag, flags *Flags, flagValue string) int {
	fn, ok := flagTypeHandlerMap[matchedFlag.flagType]
	if !ok {
		return a.stop(ErrUnsupportedFlagType, cmd, map[string]string{
			"value": fmt.Sprint(matchedFlag.flagType),
		})
	}

	if code := fn(a, cmd, matchedFlag, flags, flagValue); code != StateContinue {
		return code
	}

	return StateContinue
}

func (a *App) validateArgCount(cmd *Command, nargs int) int {
	if cmd.maxArg == 0 && cmd.minArg == 0 && nargs > 0 {
		return a.stop(ErrUnexpectedArgument, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.minArg > 0 && nargs < cmd.minArg {
		return a.stop(ErrTooFewArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.maxArg > 0 && nargs > cmd.maxArg {
		return a.stop(ErrTooManyArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}

	return StateContinue
}

func (a *App) handleHelpAndVersion(arg string, cmd *Command) int {
	switch arg {
	case "--help", "-h":
		if cmd == a.root {
			return a.stop(ErrHelp, nil, nil)
		} else {
			return a.stop(ErrCommandHelp, cmd, nil)
		}
	case "--version", "-v":
		if cmd == a.root && a.version != "" {
			return a.stop(ErrVersion, nil, nil)
		}
	}

	return StateContinue
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
