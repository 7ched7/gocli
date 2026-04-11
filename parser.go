package gocli

import (
	"fmt"
	"strings"
)

type parser struct {
	positionalOnly   bool
	helpRequested    bool
	versionRequested bool
}

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
	p := &parser{}

	ctx := &Context{
		app:     a,
		command: a.root,
		args:    []string{},
		flags:   map[string]FlagInfo{},
	}

	// Global flag values mapping
	for _, f := range ctx.command.Flags() {
		ctx.flags[f.Name()] = f
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if p.positionalOnly {
			ctx.args = append(ctx.args, arg)
			continue
		}

		// Enable positional-only mode
		if arg == "--" {
			p.positionalOnly = true
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

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, code = a.handleShortFlag(p, ctx, arg, args, i)
		} else {
			newi, code = a.handleLongFlag(p, ctx, arg, args, i)
		}

		if code != stateContinue {
			return nil, code
		}

		if code := a.handleHelpAndVersion(p, ctx.command); code != stateContinue {
			return nil, code
		}

		i = newi
	}

	cmd := ctx.command

	if cmd == a.root && cmd.action() == nil && len(ctx.args) == 0 {
		return nil, a.cliExit(MsgNoCommand, cmd, nil)
	}

	if cmd != a.root && len(cmd.Subcommands()) > 0 && cmd.action() == nil && cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
		return nil, a.cliExit(MsgSubcommandRequired, cmd, map[string]string{
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
			return a.cliExit(MsgUnknownCommand, cmd, map[string]string{
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

func (a *App) findFlag(p *parser, cmd CommandInfo, flagName string) (FlagInfo, int) {
	var matchedFlag FlagInfo

	matches := func(flagName string, f FlagInfo) bool {
		if f != nil {
			return flagName == "--"+f.Name() || (f.Alias() != "" && flagName == "-"+f.Alias())
		}
		return false
	}

	// Help flag
	f := a.config.HelpFlag
	if f != nil && matches(flagName, f) {
		matchedFlag = f
		p.helpRequested = true
	}

	// Version flag
	if matchedFlag == nil && cmd == a.root && a.version != "" {
		f := a.config.VersionFlag
		if f != nil && matches(flagName, f) {
			matchedFlag = f
			p.versionRequested = true
		}
	}

	// Local flags
	if matchedFlag == nil && cmd != a.root {
		for _, f := range cmd.Flags() {
			if matches(flagName, f) {
				matchedFlag = f
				break
			}
		}
	}

	// Global flags
	if matchedFlag == nil {
		for _, f := range a.root.flags {
			if matches(flagName, f) {
				matchedFlag = f
				break
			}
		}
	}

	if matchedFlag == nil {
		return nil, a.cliExit(MsgInvalidFlag, cmd, map[string]string{
			"flag": flagName,
		})
	}

	return matchedFlag, stateContinue
}

func (a *App) handleShortFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, int) {
	var matchedFlag FlagInfo
	var code int

	for j, f := range arg[1:] {
		matchedFlag, code = a.findFlag(p, ctx.command, "-"+string(f))
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
				return i, a.cliExit(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": matchedFlag.Alias(),
				})
			}
		}

		if code := a.handleFlagValue(ctx, matchedFlag, flagValue); code != stateContinue {
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

func (a *App) handleLongFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, int) {
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

	matchedFlag, code := a.findFlag(p, ctx.command, flagName)
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
				return i, a.cliExit(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": matchedFlag.Name(),
				})
			}
		}
	}

	if code := a.handleFlagValue(ctx, matchedFlag, flagValue); code != stateContinue {
		return i, code
	}

	return i, stateContinue
}

func (a *App) handleFlagValue(ctx *Context, matchedFlag FlagInfo, flagValue string) int {
	if err := matchedFlag.Value().Set(flagValue); err != nil {
		return a.handleFlagValueError(ctx.command, matchedFlag, flagValue, err)
	}

	if matchedFlag.role() != flagHelp && matchedFlag.role() != flagVersion {
		matchedFlag.set()
		ctx.flags[matchedFlag.Name()] = matchedFlag
	}

	return stateContinue
}

func (a *App) handleFlagValueError(cmd CommandInfo, matchedFlag FlagInfo, flagValue string, err error) int {
	errInfo := map[string]string{
		"flag":  matchedFlag.Name(),
		"value": flagValue,
	}

	switch matchedFlag.Value().(type) {
	case *typeInt:
		return a.cliExit(MsgIntParseError, cmd, errInfo)
	case *typeFloat64:
		return a.cliExit(MsgFloat64ParseError, cmd, errInfo)
	case *typeBool:
		return a.cliExit(MsgBoolParseError, cmd, errInfo)
	default:
		return a.appExit(err, exitUsage)
	}
}

func (a *App) handleHelpAndVersion(p *parser, cmd CommandInfo) int {
	if cmd == a.root {
		if p.helpRequested {
			return a.cliExit(MsgHelp, cmd, nil)
		}

		if p.versionRequested && a.config.VersionFlag != nil && a.version != "" {
			return a.cliExit(MsgVersion, cmd, nil)
		}
	} else {
		if p.helpRequested {
			return a.cliExit(MsgCommandHelp, cmd, nil)
		}
	}

	return stateContinue
}

func (a *App) validate(ctx *Context) int {
	nargs := len(ctx.args)
	cmd := ctx.command

	if cmd.MinArg() == 0 && cmd.MaxArg() == 0 && nargs > 0 {
		return a.cliExit(MsgUnexpectedArgument, cmd, map[string]string{
			"argument": ctx.args[0],
		})
	}
	if cmd.MinArg() > 0 && nargs < cmd.MinArg() {
		return a.cliExit(MsgTooFewArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.MaxArg() > 0 && nargs > cmd.MaxArg() {
		return a.cliExit(MsgTooManyArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}

	for _, f := range ctx.flags {
		if f.IsRequired() && !f.IsSet() {
			return a.cliExit(MsgFlagRequired, cmd, map[string]string{
				"flag": f.Name(),
			})
		}
	}

	for _, f := range ctx.flags {
		if f.IsSet() {
			if err := f.Validate(ctx); err != nil {
				return a.appExit(err, exitUsage)
			}
		}
	}

	return stateContinue
}

func (a *App) run(ctx *Context) int {
	if ctx.command.action() != nil {
		if err := ctx.command.action()(ctx); err != nil {
			return a.appExit(err, exitError)
		}
	}

	return stateContinue
}
