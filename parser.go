package gocli

import (
	"fmt"
	"strings"
)

func (a *App) handler(args []string) int {
	ctx, code := a.parse(args)
	if code != stateContinue {
		return code
	}

	if code = a.validate(ctx); code != stateContinue {
		return code
	}

	a.run(ctx)
	return exitOK
}

func (a *App) parse(args []string) (*Context, int) {
	cmd := a.root

	ctx := &Context{
		app:     a,
		command: cmd,
		args:    []string{},
		flags:   map[string]FlagInfo{},
	}

	positionalOnly := false

	// Global flag values mapping
	for _, f := range cmd.flags {
		ctx.flags[f.Name()] = f
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

		var code int
		var newi int

		// Positional argument
		if !strings.HasPrefix(arg, "-") {
			cmd, code = a.handleArgument(ctx, cmd, arg)
			if code != stateContinue {
				return nil, code
			}

			ctx.command = cmd
			continue
		}

		code = a.handleHelpAndVersion(arg, cmd)
		if code != stateContinue {
			return nil, code
		}

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, code = a.handleShortFlag(ctx.flags, cmd, arg, args, i)
		} else {
			newi, code = a.handleLongFlag(ctx.flags, cmd, arg, args, i)
		}

		if code != stateContinue {
			return nil, code
		}

		i = newi
	}

	if cmd == a.root && a.root.action == nil && len(ctx.args) == 0 {
		return nil, a.stop(MsgNoCommand, cmd, nil)
	}

	if cmd != a.root && len(cmd.subcommands) > 0 && cmd.action == nil && cmd.minArg == 0 && cmd.maxArg == 0 {
		return nil, a.stop(MsgSubcommandRequired, cmd, map[string]string{
			"command": cmd.name,
		})
	}

	return ctx, stateContinue
}

func (a *App) handleArgument(ctx *Context, cmd *Command, arg string) (*Command, int) {
	isCmd := false

	if len(ctx.args) == 0 {
		for _, c := range cmd.subcommands {
			if c.name == arg || c.alias == arg {
				isCmd = true
				cmd = c
			}
		}
	}

	if !isCmd {
		if ((cmd == a.root && cmd.action == nil) || cmd != a.root && len(cmd.subcommands) > 0) &&
			cmd.minArg == 0 && cmd.maxArg == 0 {
			return nil, a.stop(MsgUnknownCommand, cmd, map[string]string{
				"command": arg,
			})
		}

		ctx.args = append(ctx.args, arg)
	} else {
		// Flag values mapping
		for _, f := range cmd.flags {
			if _, ok := ctx.flags[f.Name()]; !ok {
				ctx.flags[f.Name()] = f
			}
		}
	}

	return cmd, stateContinue
}

func (a *App) findFlag(cmd *Command, flagName string) (FlagInfo, int) {
	var matchedFlag FlagInfo

	if cmd != a.root {
		for _, f := range cmd.flags {
			if flagName == "--"+f.Name() || (f.Alias() != "" && flagName == "-"+f.Alias()) {
				matchedFlag = f
				break
			}
		}
	}

	if matchedFlag == nil {
		for _, f := range a.root.flags {
			if flagName == "--"+f.Name() || (f.Alias() != "" && flagName == "-"+f.Alias()) {
				matchedFlag = f
				break
			}
		}
	}

	if matchedFlag == nil {
		return nil, a.stop(MsgInvalidFlag, cmd, map[string]string{
			"flag": flagName,
		})
	}

	return matchedFlag, stateContinue
}

func (a *App) handleShortFlag(flags map[string]FlagInfo, cmd *Command, arg string, args []string, i int) (int, int) {
	var matchedFlag FlagInfo
	var code int

	for j, f := range arg[1:] {
		matchedFlag, code = a.findFlag(cmd, "-"+string(f))
		if code != stateContinue {
			return i, code
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
				return i, a.stop(MsgFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.Alias(),
				})
			}
		}

		if code := a.handleFlagValue(flags, matchedFlag, flagValue); code != stateContinue {
			return i, code
		}

		switch matchedFlag.Value().(type) {
		case *typeBool:
			continue
		}
		break
	}

	return i, stateContinue
}

func (a *App) handleLongFlag(flags map[string]FlagInfo, cmd *Command, arg string, args []string, i int) (int, int) {
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

	matchedFlag, code := a.findFlag(cmd, flagName)
	if code != stateContinue {
		return i, code
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
				return i, a.stop(MsgFlagValueMissing, cmd, map[string]string{
					"flag": matchedFlag.Name(),
				})
			}
		}
	}

	if code := a.handleFlagValue(flags, matchedFlag, flagValue); code != stateContinue {
		return i, code
	}

	return i, stateContinue
}

func (a *App) handleFlagValue(flags map[string]FlagInfo, matchedFlag FlagInfo, flagValue string) int {
	if err := matchedFlag.Value().Set(flagValue); err != nil {
		fmt.Fprint(a.stderr, err)
		return exitUsage
	}

	matchedFlag.set()
	flags[matchedFlag.Name()] = matchedFlag

	return stateContinue
}

func (a *App) handleHelpAndVersion(arg string, cmd *Command) int {
	switch arg {
	case "--help", "-h":
		if cmd == a.root {
			return a.stop(MsgHelp, cmd, nil)
		} else {
			return a.stop(MsgCommandHelp, cmd, nil)
		}
	case "--version":
		if cmd == a.root && a.version != "" {
			return a.stop(MsgVersion, cmd, nil)
		}
	}

	return stateContinue
}

func (a *App) validate(ctx *Context) int {
	nargs := len(ctx.args)
	cmd := ctx.command

	if cmd.maxArg == 0 && cmd.minArg == 0 && nargs > 0 {
		return a.stop(MsgUnexpectedArgument, cmd, map[string]string{
			"argument": ctx.args[0],
		})
	}
	if cmd.minArg > 0 && nargs < cmd.minArg {
		return a.stop(MsgTooFewArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.maxArg > 0 && nargs > cmd.maxArg {
		return a.stop(MsgTooManyArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}

	for _, f := range ctx.flags {
		if f.IsRequired() && !f.IsSet() {
			return a.stop(MsgFlagRequired, cmd, map[string]string{
				"flag": f.Name(),
			})
		}
	}

	for _, f := range ctx.flags {
		if f.IsSet() {
			if err := f.Validate(ctx); err != nil {
				fmt.Fprint(a.stderr, err)
				return exitUsage
			}
		}
	}

	return stateContinue
}

func (a *App) run(ctx *Context) {
	if ctx.command.action != nil {
		ctx.command.action(ctx)
	}
}
