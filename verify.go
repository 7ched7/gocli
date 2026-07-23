package gocli

import (
	"fmt"
)

const (
	errCmdNameEmpty      = "command %q: command name cannot be empty"
	errCmdDuplicateName  = "command %q: duplicate command name %q"
	errCmdDuplicateAlias = "command %q: duplicate command alias %q"

	errArgNameEmpty             = "command %q: argument name cannot be empty"
	errArgAfterVariadic         = "command %q: argument %q cannot follow a variadic argument"
	errArgRequiredAfterOptional = "command %q: required argument %q cannot follow an optional argument"
	errArgDuplicateName         = "command %q: duplicate argument name %q"
	errArgMinNegative           = "command %q: minimum value of argument %q cannot be negative"
	errArgMinGreaterThanMax     = "command %q: minimum value of argument %q cannot be greater than maximum value"
	errArgMinAndMaxZero         = "command %q: minimum and maximum values for argument %q cannot both be zero"

	errFlagNameAndShorthandEmpty  = "command %q: flag name and shorthand cannot both be empty"
	errFlagShorthandCharacterLong = "command %q: flag shorthand %q cannot be longer than 1 character"
	errFlagDuplicateName          = "command %q: duplicate flag name %q"
	errFlagDuplicateShorthand     = "command %q: duplicate flag shorthand %q"

	errDefaultFlagDuplicateName      = "command %q: duplicate default flag name %q"
	errDefaultFlagDuplicateShorthand = "command %q: duplicate default flag shorthand %q"
)

func panicf(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}

func (c *Command) verifyCommands(commands ...CommandInfo) {
	seen := make(map[string]struct{})

	for _, existing := range c.subcommands {
		seen[existing.Name()] = struct{}{}
		if existing.Alias() != "" {
			seen[existing.Alias()] = struct{}{}
		}
	}

	for _, cmd := range commands {
		currName := cmd.Name()
		currAlias := cmd.Alias()

		if currName == "" {
			panicf(errCmdNameEmpty, c.name)
		}

		if _, ok := seen[currName]; ok {
			panicf(errCmdDuplicateName, c.name, currName)
		}

		if currAlias != "" {
			if _, ok := seen[currAlias]; ok {
				panicf(errCmdDuplicateAlias, c.name, currAlias)
			}
			seen[currAlias] = struct{}{}
		}

		seen[currName] = struct{}{}
	}
}

func (c *Command) verifyArguments(arguments ...ArgumentInfo) {
	seen := make(map[string]struct{})
	hasVariadic := false
	hasOptional := false

	for _, existing := range c.arguments {
		seen[existing.Name()] = struct{}{}
		if existing.IsVariadic() {
			hasVariadic = true
		}
		if !existing.IsRequired() {
			hasOptional = true
		}
	}

	for _, a := range arguments {
		currName := a.Name()

		if currName == "" {
			panicf(errArgNameEmpty, c.name)
		}

		if _, ok := seen[currName]; ok {
			panicf(errArgDuplicateName, c.name, currName)
		}

		if hasOptional && a.IsRequired() {
			panicf(errArgRequiredAfterOptional, c.name, currName)
		}
		if hasVariadic {
			panicf(errArgAfterVariadic, c.name, currName)
		}

		if !a.IsRequired() {
			hasOptional = true
		}
		if a.IsVariadic() {
			hasVariadic = true
		}

		min := a.Min()
		max := a.Max()

		if min < 0 {
			panicf(errArgMinNegative, c.name, currName)
		}

		if max != -1 && min > max {
			panicf(errArgMinGreaterThanMax, c.name, currName)
		}

		if min == 0 && max == 0 {
			panicf(errArgMinAndMaxZero, c.name, currName)
		}

		seen[currName] = struct{}{}
	}
}

func (c *Command) verifyFlags(flags ...FlagInfo) {
	seen := make(map[string]struct{})

	for _, existing := range c.flags {
		if existing.Name() != "" {
			seen[existing.Name()] = struct{}{}
		}
		if existing.Shorthand() != "" {
			seen[existing.Shorthand()] = struct{}{}
		}
	}

	for _, f := range flags {
		currName := f.Name()
		currShorthand := f.Shorthand()

		if currName == "" && currShorthand == "" {
			panicf(errFlagNameAndShorthandEmpty, c.name)
		}

		if len(currShorthand) > 1 {
			panicf(errFlagShorthandCharacterLong, c.name, currShorthand)
		}

		if currName != "" {
			if _, ok := seen[currName]; ok {
				panicf(errFlagDuplicateName, c.name, currName)
			}
			seen[currName] = struct{}{}
		}

		if currShorthand != "" {
			if _, ok := seen[currShorthand]; ok {
				panicf(errFlagDuplicateShorthand, c.name, currShorthand)
			}
			seen[currShorthand] = struct{}{}
		}
	}
}

func (a *App) verifyDefaultFlags(c CommandInfo) {
	seen := make(map[string]struct{})

	if hf := a.config.HelpFlag; hf != nil {
		if name := hf.Name(); name != "" {
			seen[name] = struct{}{}
		}

		if shorthand := hf.Shorthand(); shorthand != "" {
			seen[shorthand] = struct{}{}
		}
	}

	if c == a.root {
		if vf := a.config.VersionFlag; vf != nil {
			if name := vf.Name(); name != "" {
				if _, ok := seen[name]; ok {
					panicf(errDefaultFlagDuplicateName, c.Name(), name)
				}
				seen[name] = struct{}{}
			}

			if shorthand := vf.Shorthand(); shorthand != "" {
				if _, ok := seen[shorthand]; ok {
					panicf(errDefaultFlagDuplicateShorthand, c.Name(), shorthand)
				}
				seen[shorthand] = struct{}{}
			}
		}
	}

	for _, f := range c.Flags() {
		if name := f.Name(); name != "" {
			if _, ok := seen[name]; ok {
				panicf(errDefaultFlagDuplicateName, c.Name(), name)
			}
		}

		if shorthand := f.Shorthand(); shorthand != "" {
			if _, ok := seen[shorthand]; ok {
				panicf(errDefaultFlagDuplicateShorthand, c.Name(), shorthand)
			}
		}
	}
}
