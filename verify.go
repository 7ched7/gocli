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

func (a *App) verifyFlags(cmd CommandInfo, errs *[]error) {
	a.verifyFlag(cmd, cmd.Flags(), true, false, errs)
}

func (a *App) verifySystemFlags(errs *[]error) {
	a.verifyFlag(a.root, a.systemFlags(true), false, false, errs)
}

func (a *App) verifyGlobalFlags(errs *[]error) {
	a.verifyFlag(a.root, a.GlobalFlags(), true, true, errs)
}

func (a *App) verifyFlag(
	cmd CommandInfo,
	flags []FlagInfo,
	checkHelp bool,
	checkVersion bool,
	errs *[]error,
) {
	nameMap := make(map[string]bool)

	for _, f := range flags {
		name := strings.TrimSpace(f.Name())
		alias := strings.TrimSpace(f.Alias())

		if name == "" && alias == "" {
			*errs = append(*errs, fmt.Errorf(
				"flag name and alias cannot both be empty (command: %q)",
				cmd.Name(),
			))
			continue
		}

		if checkHelp {
			a.verifyHelpFlag(cmd, name, alias, errs)
		}

		if checkVersion {
			a.verifyVersionFlag(cmd, name, alias, errs)
		}

		if name != "" {
			if _, ok := nameMap[name]; ok {
				*errs = append(*errs, fmt.Errorf(
					"duplicate flag name %q (command %q)",
					name,
					cmd.Name(),
				))
			}

			nameMap[name] = true
		}

		if alias != "" {
			if len(alias) > 1 {
				*errs = append(*errs, fmt.Errorf(
					"flag alias %q cannot be longer than 1 character (command %q)",
					alias,
					cmd.Name(),
				))
			}

			if _, ok := nameMap[alias]; ok {
				*errs = append(*errs, fmt.Errorf(
					"duplicate flag alias %q (command %q)",
					alias,
					cmd.Name(),
				))
			}

			nameMap[alias] = true
		}
	}
}

func (a *App) verifyHelpFlag(cmd CommandInfo, name, alias string, errs *[]error) {
	helpFlag := a.config.HelpFlag

	if helpFlag != nil {
		if name != "" && helpFlag.Name() != "" && name == helpFlag.Name() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag name %q (command: %q)",
				helpFlag.Name(),
				cmd.Name(),
			))
		}

		if alias != "" && helpFlag.Alias() != "" && alias == helpFlag.Alias() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag alias %q (command: %q)",
				helpFlag.Alias(),
				cmd.Name(),
			))
		}
	}
}

func (a *App) verifyVersionFlag(cmd CommandInfo, name, alias string, errs *[]error) {
	versionFlag := a.config.VersionFlag

	if versionFlag != nil {
		if name != "" && versionFlag.Name() != "" && name == versionFlag.Name() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag name %q (command: %q)",
				versionFlag.Name(),
				cmd.Name(),
			))
		}

		if alias != "" && versionFlag.Alias() != "" && alias == versionFlag.Alias() {
			*errs = append(*errs, fmt.Errorf(
				"duplicate system flag alias %q (command: %q)",
				versionFlag.Alias(),
				cmd.Name(),
			))
		}
	}
}

func (a *App) verifyArguments(cmd CommandInfo, errs *[]error) {
	nameMap := make(map[string]bool)

	for i, arg := range cmd.Arguments() {
		name := strings.TrimSpace(arg.Name())

		if name == "" {
			*errs = append(*errs, fmt.Errorf(
				"argument name cannot be empty (command: %q)",
				cmd.Name(),
			))
			continue
		}

		if i != len(cmd.Arguments())-1 && arg.IsVariadic() {
			*errs = append(*errs, fmt.Errorf(
				"variadic argument %q must be the last argument (command: %q)",
				arg.Name(),
				cmd.Name(),
			))
		}

		if _, ok := nameMap[name]; ok {
			*errs = append(*errs, fmt.Errorf(
				"duplicate argument name %q (command: %q)",
				arg.Name(),
				cmd.Name(),
			))
		}

		if arg.Min() < 0 {
			*errs = append(*errs, fmt.Errorf(
				"minimum value of argument %q cannot be negative (command: %q)",
				arg.Name(),
				cmd.Name(),
			))
		}

		if arg.Max() != -1 && arg.Min() > arg.Max() {
			*errs = append(*errs, fmt.Errorf(
				"minimum value of argument %q cannot be greater than maximum value (command: %q)",
				arg.Name(),
				cmd.Name(),
			))
		}

		if arg.Min() == 0 && arg.Max() == 0 {
			*errs = append(*errs, fmt.Errorf(
				"minimum and maximum values for argument %q cannot both be zero (command: %q)",
				arg.Name(),
				cmd.Name(),
			))
		}

		nameMap[name] = true
	}
}
