package gocli

import (
	"fmt"
	"strings"
)

func (a *App) verify() []error {
	var errs []error
	a.verifyCommands(&errs)
	a.verifySystemFlags(&errs)
	a.verifyGlobalFlags(&errs)
	return errs
}

func (a *App) verifyCommands(errs *[]error) {
	var walk func([]CommandInfo)

	walk = func(commands []CommandInfo) {
		nameMap := make(map[string]bool)

		for _, c := range commands {
			name := strings.TrimSpace(c.Name())
			alias := strings.TrimSpace(c.Alias())

			if name == "" {
				*errs = append(*errs, fmt.Errorf(
					"command name cannot be empty (command: %q)",
					c.Parent().Name(),
				))
				continue
			}

			if _, ok := nameMap[name]; ok {
				*errs = append(*errs, fmt.Errorf(
					"duplicate command name %q (command: %q)",
					name,
					c.Parent().Name(),
				))
			}

			nameMap[name] = true

			if alias != "" {
				if _, ok := nameMap[alias]; ok {
					*errs = append(*errs, fmt.Errorf(
						"duplicate command alias %q (command: %q)",
						alias,
						c.Parent().Name(),
					))
				}

				nameMap[alias] = true
			}

			a.verifyFlags(c, errs)
			a.verifyArguments(c, errs)

			walk(c.Subcommands())
		}
	}

	walk(a.root.subcommands)
}

func (a *App) verifyFlags(c CommandInfo, errs *[]error) {
	a.verifyFlag(c, c.Flags(), true, false, errs)
}

func (a *App) verifySystemFlags(errs *[]error) {
	a.verifyFlag(a.root, a.systemFlags(true), false, false, errs)
}

func (a *App) verifyGlobalFlags(errs *[]error) {
	a.verifyFlag(a.root, a.GlobalFlags(), true, true, errs)
}

func (a *App) verifyFlag(
	c CommandInfo,
	flags []FlagInfo,
	checkHelp bool,
	checkVersion bool,
	errs *[]error,
) {
	nameMap := make(map[string]bool)

	for _, f := range flags {
		name := strings.TrimSpace(f.Name())
		shorthand := strings.TrimSpace(f.Shorthand())

		if name == "" && shorthand == "" {
			*errs = append(*errs, fmt.Errorf(
				"flag name and shorthand cannot both be empty (command: %q)",
				c.Name(),
			))
			continue
		}

		if checkHelp {
			a.verifyHelpFlag(c, name, shorthand, errs)
		}

		if checkVersion {
			a.verifyVersionFlag(c, name, shorthand, errs)
		}

		if name != "" {
			if _, ok := nameMap[name]; ok {
				*errs = append(*errs, fmt.Errorf(
					"duplicate flag name %q (command %q)",
					name,
					c.Name(),
				))
			}

			nameMap[name] = true
		}

		if shorthand != "" {
			if len(shorthand) > 1 {
				*errs = append(*errs, fmt.Errorf(
					"flag shorthand %q cannot be longer than 1 character (command %q)",
					shorthand,
					c.Name(),
				))
			}

			if _, ok := nameMap[shorthand]; ok {
				*errs = append(*errs, fmt.Errorf(
					"duplicate flag shorthand %q (command %q)",
					shorthand,
					c.Name(),
				))
			}

			nameMap[shorthand] = true
		}
	}
}

func (a *App) verifyHelpFlag(c CommandInfo, name, shorthand string, errs *[]error) {
	f := a.config.HelpFlag

	if f != nil {
		if name != "" && f.Name() != "" && name == f.Name() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag name %q (command: %q)",
				f.Name(),
				c.Name(),
			))
		}

		if shorthand != "" && f.Shorthand() != "" && shorthand == f.Shorthand() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag shorthand %q (command: %q)",
				f.Shorthand(),
				c.Name(),
			))
		}
	}
}

func (a *App) verifyVersionFlag(c CommandInfo, name, shorthand string, errs *[]error) {
	f := a.config.VersionFlag

	if f != nil {
		if name != "" && f.Name() != "" && name == f.Name() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag name %q (command: %q)",
				f.Name(),
				c.Name(),
			))
		}

		if shorthand != "" && f.Shorthand() != "" && shorthand == f.Shorthand() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag shorthand %q (command: %q)",
				f.Shorthand(),
				c.Name(),
			))
		}
	}
}

func (a *App) verifyArguments(c CommandInfo, errs *[]error) {
	nameMap := make(map[string]bool)

	for i, arg := range c.Arguments() {
		name := strings.TrimSpace(arg.Name())

		if name == "" {
			*errs = append(*errs, fmt.Errorf(
				"argument name cannot be empty (command: %q)",
				c.Name(),
			))
			continue
		}

		if i != len(c.Arguments())-1 && arg.IsVariadic() {
			*errs = append(*errs, fmt.Errorf(
				"variadic argument %q must be the last argument (command: %q)",
				arg.Name(),
				c.Name(),
			))
		}

		if _, ok := nameMap[name]; ok {
			*errs = append(*errs, fmt.Errorf(
				"duplicate argument name %q (command: %q)",
				arg.Name(),
				c.Name(),
			))
		}

		if arg.Min() < 0 {
			*errs = append(*errs, fmt.Errorf(
				"minimum value of argument %q cannot be negative (command: %q)",
				arg.Name(),
				c.Name(),
			))
		}

		if arg.Max() != -1 && arg.Min() > arg.Max() {
			*errs = append(*errs, fmt.Errorf(
				"minimum value of argument %q cannot be greater than maximum value (command: %q)",
				arg.Name(),
				c.Name(),
			))
		}

		if arg.Min() == 0 && arg.Max() == 0 {
			*errs = append(*errs, fmt.Errorf(
				"minimum and maximum values for argument %q cannot both be zero (command: %q)",
				arg.Name(),
				c.Name(),
			))
		}

		nameMap[name] = true
	}
}
