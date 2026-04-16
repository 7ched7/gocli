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

func (a *App) handler(args []string) error {
	ctx, err := a.parse(args)
	if err != nil {
		return err
	}

	if err := a.validate(ctx); err != nil {
		return err
	}

	if err := a.run(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) parse(args []string) (*Context, error) {
	p := &parser{}

	ctx := &Context{
		app:     a,
		command: a.root,
		args:    []string{},
		flags:   map[string]FlagInfo{},
	}

	// Global flag values mapping
	for _, f := range ctx.command.Flags() {
		ctx.flags[flagDisplayName(f, false)] = f
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

		var err error
		var newi int

		// Positional argument
		if !strings.HasPrefix(arg, "-") {
			if err := a.handleArgument(ctx, arg); err != nil {
				return nil, err
			}
			continue
		}

		if !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			newi, err = a.handleShortFlag(p, ctx, arg, args, i)
		} else {
			newi, err = a.handleLongFlag(p, ctx, arg, args, i)
		}

		if err != nil {
			return nil, err
		}

		if err := a.handleHelpAndVersion(p, ctx.command); err != nil {
			return nil, err
		}

		i = newi
	}

	cmd := ctx.command

	if cmd == a.root && cmd.action() == nil && len(ctx.args) == 0 && cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
		return nil, a.exitWithMsg(MsgNoCommand, cmd, nil)
	}

	if cmd != a.root && len(cmd.Subcommands()) > 0 && cmd.action() == nil && cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
		return nil, a.exitWithMsg(MsgSubcommandRequired, cmd, map[string]string{
			"command": commandDisplayName(cmd),
		})
	}

	return ctx, nil
}

func (a *App) handleArgument(ctx *Context, arg string) error {
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
			return a.exitWithMsg(MsgUnknownCommand, cmd, map[string]string{
				"command": arg,
			})
		}

		ctx.args = append(ctx.args, arg)
	} else {
		// Flag values mapping
		for _, f := range cmd.Flags() {
			if _, ok := ctx.flags[flagDisplayName(f, false)]; !ok {
				ctx.flags[flagDisplayName(f, false)] = f
			}
		}
	}

	return nil
}

func (a *App) findFlag(p *parser, cmd CommandInfo, flagName string) (FlagInfo, error) {
	var matchedFlag FlagInfo

	matches := func(flagName string, f FlagInfo) bool {
		if f != nil {
			return (f.Name() != "" && flagName == "--"+f.Name()) || (f.Alias() != "" && flagName == "-"+f.Alias())
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
	if matchedFlag == nil && cmd == a.root {
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
		return nil, a.exitWithMsg(MsgInvalidFlag, cmd, map[string]string{
			"flag": flagName,
		})
	}

	return matchedFlag, nil
}

func (a *App) handleShortFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, error) {
	var matchedFlag FlagInfo
	var err error

	for j, f := range arg[1:] {
		matchedFlag, err = a.findFlag(p, ctx.command, "-"+string(f))
		if err != nil {
			return i, err
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
				return i, a.exitWithMsg(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": flagDisplayName(matchedFlag, true),
				})
			}
		}

		if err := a.handleFlagValue(ctx, matchedFlag, flagValue); err != nil {
			return i, err
		}

		switch matchedFlag.Value().(type) {
		case *typeBool:
			continue
		}
		break
	}

	return i, nil
}

func (a *App) handleLongFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, error) {
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

	matchedFlag, err := a.findFlag(p, ctx.command, flagName)
	if err != nil {
		return i, err
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
				return i, a.exitWithMsg(MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": flagDisplayName(matchedFlag, true),
				})
			}
		}
	}

	if err := a.handleFlagValue(ctx, matchedFlag, flagValue); err != nil {
		return i, err
	}

	return i, nil
}

func (a *App) handleFlagValue(ctx *Context, matchedFlag FlagInfo, flagValue string) error {
	if err := matchedFlag.Value().Set(flagValue); err != nil {
		return a.handleFlagValueError(ctx.command, matchedFlag, flagValue, err)
	}

	if matchedFlag.role() != flagHelp && matchedFlag.role() != flagVersion {
		matchedFlag.set()
		ctx.flags[flagDisplayName(matchedFlag, false)] = matchedFlag
	}

	return nil
}

func (a *App) handleFlagValueError(cmd CommandInfo, matchedFlag FlagInfo, flagValue string, err error) error {
	errInfo := map[string]string{
		"flag":  flagDisplayName(matchedFlag, true),
		"value": flagValue,
	}

	switch matchedFlag.Value().(type) {
	case *typeInt:
		return a.exitWithMsg(MsgIntParseError, cmd, errInfo)
	case *typeFloat64:
		return a.exitWithMsg(MsgFloat64ParseError, cmd, errInfo)
	case *typeBool:
		return a.exitWithMsg(MsgBoolParseError, cmd, errInfo)
	default:
		return a.exitWithErr(err, exitUsage)
	}
}

func (a *App) handleHelpAndVersion(p *parser, cmd CommandInfo) error {
	if cmd == a.root {
		if p.helpRequested {
			return a.exitWithMsg(MsgHelp, cmd, nil)
		}

		if p.versionRequested {
			return a.exitWithMsg(MsgVersion, cmd, nil)
		}
	} else {
		if p.helpRequested {
			return a.exitWithMsg(MsgCommandHelp, cmd, nil)
		}
	}

	return nil
}

func (a *App) validate(ctx *Context) error {
	nargs := len(ctx.args)
	cmd := ctx.command

	if cmd.MinArg() == 0 && cmd.MaxArg() == 0 && nargs > 0 {
		return a.exitWithMsg(MsgUnexpectedArgument, cmd, map[string]string{
			"argument": ctx.args[0],
		})
	}
	if cmd.MinArg() > 0 && nargs < cmd.MinArg() {
		return a.exitWithMsg(MsgTooFewArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}
	if cmd.MaxArg() > 0 && nargs > cmd.MaxArg() {
		return a.exitWithMsg(MsgTooManyArguments, cmd, map[string]string{
			"number": fmt.Sprint(nargs),
		})
	}

	for _, f := range ctx.flags {
		if f.IsRequired() && !f.IsSet() {
			return a.exitWithMsg(MsgFlagRequired, cmd, map[string]string{
				"flag": flagDisplayName(f, true),
			})
		}
	}

	for _, f := range ctx.flags {
		if f.IsSet() {
			if err := f.Validate(ctx); err != nil {
				return a.exitWithErr(err, exitUsage)
			}
		}
	}

	return nil
}

func (a *App) run(ctx *Context) error {
	if ctx.command.action() != nil {
		if err := ctx.command.action()(ctx); err != nil {
			return a.exitWithErr(err, exitError)
		}
	}

	return nil
}
