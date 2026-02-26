package gocli

import (
	"fmt"
	"strings"
)

func (a *App) parseCommand(args []string) (*Context, *Command, int) {
	cmd := a.root
	parsedFlags := []FlagInfo{}

	ctx := &Context{
		app:     a,
		command: cmd,
		args:    []string{},
		flags:   map[string]FlagValue{},
	}

	positionalOnly := false

	// Default global flag values mapping
	for _, flag := range a.globalFlags {
		ctx.flags[flag.Name()] = flag.DefaultValue()
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if positionalOnly {
			ctx.args = append(ctx.args, arg)
			continue
		}

		// Enable positional-only mode
		if arg == "--" {
			positionalOnly = true
			continue
		}

		var matchedFlag FlagInfo
		var code int
		var newi int

		// Positional argument
		if !strings.HasPrefix(arg, "-") {
			cmd, code = a.handleArgument(ctx, cmd, arg)
			if code != StateContinue {
				return nil, nil, code
			}

			ctx.command = cmd
			continue
		}

		code = a.handleHelpAndVersion(arg, cmd)
		if code != StateContinue {
			return nil, nil, code
		}

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			matchedFlag, newi, code = a.handleShortFlag(ctx, cmd, arg, args, i)
		} else {
			matchedFlag, newi, code = a.handleLongFlag(ctx, cmd, arg, args, i)
		}

		if code != StateContinue {
			return nil, nil, code
		}

		parsedFlags = append(parsedFlags, matchedFlag)

		i = newi
	}

	if cmd == a.root && a.root.action == nil && len(ctx.args) == 0 {
		return nil, nil, a.stop(ErrNoCommand, nil, nil)
	}

	if len(cmd.subcommands) > 0 && cmd.minArg == 0 && cmd.maxArg == 0 {
		return nil, nil, a.stop(ErrSubcommandRequired, cmd, map[string]string{
			"command": cmd.name,
		})
	}

	if code := a.validateArgCount(cmd, ctx.args); code != StateContinue {
		return nil, nil, code
	}

	if code := a.validateFlags(ctx, parsedFlags); code != StateContinue {
		return nil, nil, code
	}

	return ctx, cmd, StateContinue
}

func (a *App) handleArgument(ctx *Context, cmd *Command, arg string) (*Command, int) {
	isCmd := false

	if len(ctx.args) == 0 {
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
		if ((cmd == a.root && cmd.action == nil) || len(cmd.subcommands) > 0) &&
			cmd.minArg == 0 && cmd.maxArg == 0 {
			return nil, a.stop(ErrUnknownCommand, cmd, map[string]string{
				"command": arg,
			})
		}

		ctx.args = append(ctx.args, arg)
	} else {
		// Default command flag values mapping
		for _, flag := range cmd.flags {
			if _, ok := ctx.flags[flag.Name()]; !ok {
				ctx.flags[flag.Name()] = flag.DefaultValue()
			}
		}
	}

	return cmd, StateContinue
}

func (a *App) findFlagName(cmd *Command, flagName string) (FlagInfo, int) {
	var matchedFlag FlagInfo

	if cmd != a.root {
		for i := range cmd.flags {
			flag := cmd.flags[i]
			if flagName == "--"+flag.Name() || (flag.Alias() != "" && flagName == "-"+flag.Alias()) {
				matchedFlag = flag
				break
			}
		}
	}

	if matchedFlag == nil {
		for i := range a.globalFlags {
			flag := a.globalFlags[i]
			if flagName == "--"+flag.Name() || (flag.Alias() != "" && flagName == "-"+flag.Alias()) {
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

func (a *App) handleLongFlag(ctx *Context, cmd *Command, arg string, args []string, i int) (FlagInfo, int, int) {
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
		return nil, i, code
	}

	switch matchedFlag.Value().(type) {
	case *typeBool:
		if flagValue == "" && !hasEqualSign {
			flagValue = "true"
		}
	default:
		if flagValue == "" {
			if i+1 < len(args) && !hasEqualSign { // --flag value
				flagValue = args[i+1]
				i++
			} else {
				return nil, i, a.stop(ErrFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.Name(),
				})
			}
		}
	}

	if code := a.handleFlagValue(ctx, matchedFlag, flagValue); code != StateContinue {
		return nil, i, code
	}

	return matchedFlag, i, StateContinue
}

func (a *App) handleShortFlag(ctx *Context, cmd *Command, arg string, args []string, i int) (FlagInfo, int, int) {
	var matchedFlag FlagInfo
	var code int

	for j, f := range arg[1:] {
		matchedFlag, code = a.findFlagName(cmd, "-"+string(f))
		if code != StateContinue {
			return nil, i, code
		}

		var flagValue string

		switch matchedFlag.Value().(type) {
		case *typeBool:
			flagValue = "true"
		default:
			if j < len(arg[1:])-1 { // -fvalue
				flagValue = arg[j+2:]
			} else if i+1 < len(args) { // -f value
				flagValue = args[i+1]
				i++
			} else {
				return nil, i, a.stop(ErrFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.Alias(),
				})
			}
		}

		if code := a.handleFlagValue(ctx, matchedFlag, flagValue); code != StateContinue {
			return nil, i, code
		}

		switch matchedFlag.Value().(type) {
		case *typeBool:
			continue
		}
		break
	}

	return matchedFlag, i, StateContinue
}

func (a *App) handleFlagValue(ctx *Context, matchedFlag FlagInfo, flagValue string) int {
	if err := matchedFlag.Value().Set(flagValue); err != nil {
		fmt.Fprint(a.stderr, err)
		return ExitUsage
	}

	ctx.flags[matchedFlag.Name()] = matchedFlag.Value()

	return StateContinue
}

func (a *App) validateArgCount(cmd *Command, args []string) int {
	nargs := len(args)

	if cmd.maxArg == 0 && cmd.minArg == 0 && nargs > 0 {
		return a.stop(ErrUnexpectedArgument, cmd, map[string]string{
			"argument": args[0],
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

func (a *App) validateFlags(ctx *Context, parsedFlags []FlagInfo) int {
	for _, f := range parsedFlags {
		if err := f.Validate(ctx); err != nil {
			fmt.Fprint(a.stderr, err)
			return ExitUsage
		}
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

func (a *App) runCommand(ctx *Context, cmd *Command) {
	if cmd.action != nil {
		cmd.action(ctx)
	}
}
