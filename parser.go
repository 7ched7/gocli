package gocli

import (
	"fmt"
	"strings"
)

type parser struct {
	rawArgs          []string
	positionalOnly   bool
	currentCommand   string
	currentFlag      string
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
		args:    map[string]ArgumentInfo{},
		flags:   map[string]FlagInfo{},
	}

	// Global flag values mapping
	for _, f := range ctx.command.Flags() {
		ctx.flags[flagDisplayName(f, false)] = f
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if p.positionalOnly {
			p.rawArgs = append(p.rawArgs, arg)
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
			if err := a.handleArgument(p, ctx, arg); err != nil {
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

		i = newi
	}

	if err := a.validateArguments(p, ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (a *App) handleArgument(p *parser, ctx *Context, arg string) error {
	hasNoArgs := len(p.rawArgs) == 0
	isCmd := false

	if hasNoArgs {
		for _, c := range ctx.command.Subcommands() {
			if (c.Name() != "" && c.Name() == arg) ||
				(c.Alias() != "" && c.Alias() == arg) {
				ctx.command = c
				p.currentCommand = arg
				isCmd = true
			}
		}
	}

	c := ctx.command

	if !isCmd {
		if hasNoArgs && len(c.Subcommands()) > 0 && len(c.Arguments()) == 0 {
			return a.exitWithMsg(exitUsage, MsgUnknownCommand, c, map[string]string{
				"command": arg,
			})
		}

		p.rawArgs = append(p.rawArgs, arg)
	} else {
		// Flag values mapping
		for _, f := range c.Flags() {
			if _, ok := ctx.flags[flagDisplayName(f, false)]; !ok {
				ctx.flags[flagDisplayName(f, false)] = f
			}
		}
	}

	return nil
}

func (a *App) findFlag(p *parser, c CommandInfo, name string) (FlagInfo, error) {
	var matched FlagInfo

	matches := func(name string, f FlagInfo) bool {
		return (f.Name() != "" && "--"+f.Name() == name) ||
			(f.Alias() != "" && "-"+f.Alias() == name)
	}

	// Help flag
	f := a.config.HelpFlag
	if f != nil && matches(name, f) {
		matched = f
		p.helpRequested = true
	}

	// Version flag
	if matched == nil && c == a.root {
		f := a.config.VersionFlag
		if f != nil && matches(name, f) {
			matched = f
			p.versionRequested = true
		}
	}

	// Local flags
	if matched == nil && c != a.root {
		for _, f := range c.Flags() {
			if matches(name, f) {
				matched = f
				break
			}
		}
	}

	// Global flags
	if matched == nil {
		for _, f := range a.root.flags {
			if matches(name, f) {
				matched = f
				break
			}
		}
	}

	if matched == nil {
		return nil, a.exitWithMsg(exitUsage, MsgInvalidFlag, c, map[string]string{
			"flag": name,
		})
	}

	p.currentFlag = name
	return matched, nil
}

func (a *App) handleShortFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, error) {
	for j, f := range arg[1:] {
		matched, err := a.findFlag(p, ctx.command, "-"+string(f))
		if err != nil {
			return i, err
		}

		var value string

		if v, ok := matched.Value().(NoArgFlag); ok {
			if err := v.SetNoArg(); err != nil {
				return i, a.handleFlagValueError(p, ctx.command, matched, value, err)
			}
		} else {
			if j < len(arg[1:])-1 { // -fvalue
				value = arg[j+2:]
			} else if i+1 < len(args) { // -f value
				value = args[i+1]
				i++
			} else {
				return i, a.exitWithMsg(exitUsage, MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": p.currentFlag,
				})
			}

			if err := matched.Value().Set(value); err != nil {
				return i, a.handleFlagValueError(p, ctx.command, matched, value, err)
			}
		}

		if err := a.handleSystemFlag(p, ctx.command); err != nil {
			return i, err
		}

		registerFlag(ctx, matched)

		switch matched.Value().(type) {
		case NoArgFlag:
			continue
		}
		break
	}

	return i, nil
}

func (a *App) handleLongFlag(p *parser, ctx *Context, arg string, args []string, i int) (int, error) {
	var name string
	var value string
	hasEqualSign := strings.Contains(arg, "=")

	if hasEqualSign {
		parts := strings.SplitN(arg, "=", 2)
		name = parts[0]
		value = parts[1]
	} else {
		name = arg
	}

	matched, err := a.findFlag(p, ctx.command, name)
	if err != nil {
		return i, err
	}

	if hasEqualSign { // --flag=value
		if err := matched.Value().Set(value); err != nil {
			return i, a.handleFlagValueError(p, ctx.command, matched, value, err)
		}
	} else {
		if v, ok := matched.Value().(NoArgFlag); ok {
			if err := v.SetNoArg(); err != nil {
				return i, a.handleFlagValueError(p, ctx.command, matched, value, err)
			}
		} else {
			if i+1 < len(args) { // --flag value
				value = args[i+1]

				if err := matched.Value().Set(value); err != nil {
					return i, a.handleFlagValueError(p, ctx.command, matched, value, err)
				}

				i++
			} else {
				return i, a.exitWithMsg(exitUsage, MsgFlagValueMissing, ctx.command, map[string]string{
					"flag": p.currentFlag,
				})
			}
		}
	}

	if err := a.handleSystemFlag(p, ctx.command); err != nil {
		return i, err
	}

	registerFlag(ctx, matched)

	return i, nil
}

func (a *App) handleFlagValueError(p *parser, c CommandInfo, matched FlagInfo, value string, err error) error {
	data := map[string]string{
		"flag":  p.currentFlag,
		"value": value,
	}

	switch matched.Value().(type) {
	case *typeInt:
		return a.exitWithMsg(exitUsage, MsgIntParseError, c, data)
	case *typeFloat64:
		return a.exitWithMsg(exitUsage, MsgFloat64ParseError, c, data)
	case *typeBool:
		return a.exitWithMsg(exitUsage, MsgBoolParseError, c, data)
	default:
		return a.exitWithErr(exitUsage, err, c)
	}
}

func registerFlag(ctx *Context, matched FlagInfo) {
	matched.set()
	ctx.flags[flagDisplayName(matched, false)] = matched
}

func (a *App) handleSystemFlag(p *parser, c CommandInfo) error {
	if c == a.root {
		if p.helpRequested {
			return a.exitWithMsg(exitOK, MsgHelp, c, nil)
		}

		if p.versionRequested {
			return a.exitWithMsg(exitOK, MsgVersion, c, nil)
		}
	} else {
		if p.helpRequested {
			return a.exitWithMsg(exitOK, MsgCommandHelp, c, nil)
		}
	}

	return nil
}

func (a *App) validateArguments(p *parser, ctx *Context) error {
	c := ctx.command
	hasNoArgs := len(p.rawArgs) == 0
	hasNoActionOrArgs := c.Action() == nil && len(c.Arguments()) == 0

	if c == a.root && hasNoArgs && hasNoActionOrArgs {
		return a.exitWithMsg(exitOK, MsgNoCommand, c, nil)
	}

	if c != a.root && hasNoArgs && hasNoActionOrArgs && len(c.Subcommands()) > 0 {
		return a.exitWithMsg(exitUsage, MsgSubcommandRequired, c, map[string]string{
			"command": p.currentCommand,
		})
	}

	if !hasNoArgs && len(c.Arguments()) == 0 {
		return a.exitWithMsg(exitUsage, MsgUnexpectedArgument, c, map[string]string{
			"argument": p.rawArgs[0],
		})
	}

	if len(c.Arguments()) == 0 {
		return nil
	}

	for i, arg := range c.Arguments() {
		hasArg := i < len(p.rawArgs)
		remaining := max(len(p.rawArgs)-i, 0)

		data := map[string]string{
			"name": arg.Name(),
			"min":  fmt.Sprint(arg.Min()),
			"max":  fmt.Sprint(arg.Max()),
			"got":  fmt.Sprint(remaining),
		}

		if arg.IsVariadic() {
			if arg.IsRequired() && arg.Min() > remaining {
				return a.exitWithMsg(exitUsage, MsgTooFewArguments, c, data)
			}

			if arg.Max() >= 0 && arg.Max() < remaining {
				return a.exitWithMsg(exitUsage, MsgTooManyArguments, c, data)
			}

			if hasArg {
				arg.set(p.rawArgs[i:]...)
			}

			ctx.args[arg.Name()] = arg
			break
		}

		if arg.IsRequired() && !hasArg {
			return a.exitWithMsg(exitUsage, MsgTooFewArguments, c, data)
		}

		if i == len(c.Arguments())-1 && remaining > 1 {
			return a.exitWithMsg(exitUsage, MsgTooManyArguments, c, data)
		}

		if hasArg {
			arg.set(p.rawArgs[i])
		}

		ctx.args[arg.Name()] = arg
	}

	return nil
}

func (a *App) validate(ctx *Context) error {
	for _, f := range ctx.flags {
		if f.IsRequired() && !f.IsSet() {
			return a.exitWithMsg(exitUsage, MsgFlagRequired, ctx.command, map[string]string{
				"flag": flagDisplayName(f, true),
			})
		}
	}

	for _, f := range ctx.flags {
		if f.IsSet() {
			if err := f.Validate(ctx); err != nil {
				return a.exitWithErr(exitUsage, err, ctx.command)
			}
		}
	}

	return nil
}

func (a *App) run(ctx *Context) error {
	if ctx.command.Action() != nil {
		if err := ctx.command.Action()(ctx); err != nil {
			return a.exitWithErr(exitError, err, ctx.command)
		}
	}

	return nil
}
