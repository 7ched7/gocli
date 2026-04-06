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

	if code = a.run(ctx); code != stateContinue {
		return code
	}

	return exitOK
}

func (a *App) parse(args []string) (*Context, int) {
	ctx := &Context{
		app:     a,
		command: a.root,
		args:    []string{},
		flags:   map[string]FlagInfo{},
	}

	positionalOnly := false

	// Global flag values mapping
	for _, f := range ctx.command.Flags() {
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
			if code = a.handleArgument(ctx, arg); code != stateContinue {
				return nil, code
			}

			continue
		}

		code = a.handleHelpAndVersion(arg, ctx.command)
		if code != stateContinue {
			return nil, code
		}

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, code = a.handleShortFlag(ctx, arg, args, i)
		} else {
			newi, code = a.handleLongFlag(ctx, arg, args, i)
		}

		if code != stateContinue {
			return nil, code
		}

		i = newi
	}

	cmd := ctx.command

	if cmd == a.root && cmd.action() == nil && len(ctx.args) == 0 {
		return nil, a.stop(MsgNoCommand, cmd, nil)
	}

	if cmd != a.root && len(cmd.Subcommands()) > 0 && cmd.action() == nil && cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
		return nil, a.stop(MsgSubcommandRequired, cmd, map[string]string{
			"command": cmd.Name(),
		})
	}

	return ctx, stateContinue
}

func (a *App) handleArgument(ctx *Context, arg string) int {
	isCmd := false

	if len(ctx.args) == 0 {
		for _, c := range ctx.command.Subcommands() {
			if c.Name() == arg || c.Alias() == arg {
				isCmd = true
				ctx.command = c
			}
		}
	}

	cmd := ctx.command

	if !isCmd {
		if ((cmd == a.root && cmd.action() == nil) || cmd != a.root && len(cmd.Subcommands()) > 0) &&
			cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
			return a.stop(MsgUnknownCommand, cmd, map[string]string{
				"command": arg,
			})
		}

		ctx.args = append(ctx.args, arg)
	} else {
		// Flag values mapping
		for _, f := range cmd.Flags() {
			if _, ok := ctx.flags[f.Name()]; !ok {
				ctx.flags[f.Name()] = f
			}
		}
	}

	return stateContinue
}

func (a *App) findFlag(cmd CommandInfo, flagName string) (FlagInfo, int) {
	var matchedFlag FlagInfo

	if cmd != a.root {
		for _, f := range cmd.Flags() {
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

func (a *App) handleShortFlag(ctx *Context, arg string, args []string, i int) (int, int) {
	var matchedFlag FlagInfo
	var code int

	for j, f := range arg[1:] {
		matchedFlag, code = a.findFlag(ctx.command, "-"+string(f))
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
				return i, a.stop(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": matchedFlag.Alias(),
				})
			}
		}

		if code := a.handleFlagValue(ctx.flags, matchedFlag, flagValue); code != stateContinue {
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

func (a *App) handleLongFlag(ctx *Context, arg string, args []string, i int) (int, int) {
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

	matchedFlag, code := a.findFlag(ctx.command, flagName)
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
				return i, a.stop(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": matchedFlag.Name(),
				})
			}
		}
	}

	if code := a.handleFlagValue(ctx.flags, matchedFlag, flagValue); code != stateContinue {
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

func (a *App) handleHelpAndVersion(arg string, cmd CommandInfo) int {
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

	if cmd.MinArg() == 0 && cmd.MaxArg() == 0 && nargs > 0 {
		return a.stop(MsgUnexpectedArgument, cmd, map[string]string{
			"argument": ctx.args[0],
		})
	}
	if cmd.MinArg() > 0 && nargs < cmd.MinArg() {
		return a.stop(MsgTooFewArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.MaxArg() > 0 && nargs > cmd.MaxArg() {
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

func (a *App) run(ctx *Context) int {
	if ctx.command.action() != nil {
		if err := ctx.command.action()(ctx); err != nil {
			fmt.Fprint(a.stderr, err)
			return exitError
		}
	}

	return stateContinue
}
