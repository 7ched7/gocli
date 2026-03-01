package gocli

import (
	"fmt"
	"strings"
)

// Help generates and returns the global help menu for the application.
// It includes usage instructions, a list of commands, global flags,
// and is displayed when the user types --help or -h or when no command is provided.
func (a *App) Help() string {
	var sb strings.Builder
	maxLen := getMaxLength(a.commands, a.globalFlags, true)

	// Usage
	fmt.Fprintf(&sb, "Usage:\n  "+a.name)

	if len(a.globalFlags) > 0 {
		fmt.Fprintf(&sb, " [global flags]")
	}
	if len(a.commands) > 0 {
		fmt.Fprintf(&sb, " [command]")
	}
	fmt.Fprintf(&sb, "\n")

	// Description
	if a.description != "" {
		appDescPart := wrapText(a.description, 0)
		fmt.Fprintf(&sb, "\n%s\n", appDescPart)
	}

	// Commands
	if len(a.commands) > 0 {
		fmt.Fprintf(&sb, "\nCommands:\n")
		sb.WriteString(writeCommands(a.commands, maxLen))
	}

	// Global flags
	fmt.Fprintf(&sb, "\nGlobal flags:\n")
	if len(a.globalFlags) > 0 {
		sb.WriteString(writeFlags(a.globalFlags, maxLen))
		fmt.Fprintf(&sb, "\n")
	}

	// Help and version flags
	hPart := fmt.Sprintf("  %s --%s", "-h,", "help")
	fmt.Fprintf(&sb, "%-*s  %s\n", maxLen, hPart, "Show help")

	if a.version != "" {
		vPart := fmt.Sprintf("  %s --%s", "-v,", "version")
		fmt.Fprintf(&sb, "%-*s  %s\n", maxLen, vPart, "Show version")
	}

	// Footer
	if len(a.commands) > 0 {
		fmt.Fprintf(&sb, "\nUse \"%s [command] --help\" for more information about a command.\n", a.name)
	}

	return sb.String()
}

// CommandHelp generates and returns a help menu for a specific command.
// It includes the full command path, argument expectations, registered subcommands/flags,
// and is displayed when the user types "[command] --help" or "[command] -h".
func (a *App) CommandHelp(cmd *Command) string {
	var sb strings.Builder
	maxLen := getMaxLength(cmd.subcommands, cmd.flags, false)

	// Usage
	fmt.Fprintf(&sb, "Usage:\n  "+a.name)
	fmt.Fprintf(&sb, "%s", cmd.fullPath())

	if len(cmd.subcommands) > 0 {
		fmt.Fprint(&sb, " [command]")
	}
	if len(cmd.flags) > 0 {
		fmt.Fprint(&sb, " [flags]")
	}

	if cmd.maxArg == 1 {
		fmt.Fprint(&sb, " [arg]")
	} else if cmd.maxArg > 1 {
		fmt.Fprintf(&sb, " [arg1...arg%d]", cmd.maxArg)
	} else if cmd.minArg > 0 {
		fmt.Fprint(&sb, " [args...]")
	}
	fmt.Fprintf(&sb, "\n")

	// Description
	if cmd.long != "" {
		longPart := wrapText(cmd.long, 0)
		fmt.Fprintf(&sb, "\n%s\n", longPart)
	}

	// Commands
	if len(cmd.subcommands) > 0 {
		fmt.Fprintf(&sb, "\nCommands:\n")
		sb.WriteString(writeCommands(cmd.subcommands, maxLen))
	}

	// Flags
	if len(cmd.flags) > 0 {
		fmt.Fprintf(&sb, "\nFlags:\n")
		sb.WriteString(writeFlags(cmd.flags, maxLen))
	}

	return sb.String()
}

func writeCommands(cmds []*Command, keyWidth int) string {
	var sb strings.Builder
	for _, cmd := range cmds {
		namePart := "  " + cmd.name
		if cmd.alias != "" {
			namePart += ", " + cmd.alias
		}
		descPart := wrapText(cmd.short, keyWidth+2)
		fmt.Fprintf(&sb, "%-*s  %s\n", keyWidth, namePart, descPart)
	}
	return sb.String()
}

func writeFlags(flags []FlagInfo, keyWidth int) string {
	var sb strings.Builder
	for _, f := range flags {
		aliasPart := "   "
		if f.Alias() != "" {
			aliasPart = "-" + f.Alias() + ","
		}

		defaultPart := ""
		if f.DefaultValue().String() != zeroValues[f.HelpType()] {
			defaultPart = fmt.Sprintf("(default: %s)", f.DefaultValue().String())
		}

		helpType := f.HelpType()
		if helpType == "bool" {
			helpType = ""
		}

		flagPart := fmt.Sprintf("  %s --%s %s", aliasPart, f.Name(), helpType)
		descPart := wrapText(f.Description()+" "+defaultPart, keyWidth+2)

		fmt.Fprintf(&sb, "%-*s  %s\n", keyWidth, flagPart, descPart)
	}
	return sb.String()
}

func getMaxLength(cmds []*Command, flags []FlagInfo, isRoot bool) int {
	maxLen := 0
	padding := 2
	maxWidth := 25
	aliasLen := 3 // -x,

	if isRoot {
		maxLen = 15
	}

	for _, cmd := range cmds {
		l := len(cmd.name) + len(cmd.alias)
		if l > maxLen {
			maxLen = l
		}
	}

	for _, f := range flags {
		l := 2 + aliasLen + 1 + len(f.Name()) + 1 + len(f.HelpType())
		if l > maxLen {
			maxLen = l
		}
	}

	if maxLen > maxWidth {
		maxLen = maxWidth
	}

	return maxLen + padding
}

func wrapText(text string, keyWidth int) string {
	tWidth := 80

	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	var currLine strings.Builder
	indent := strings.Repeat(" ", keyWidth)

	for _, word := range words {
		// Start a new line if it exceeds width
		if currLine.Len()+len(word)+1 > tWidth-keyWidth {
			result.WriteString(currLine.String() + "\n" + indent)
			currLine.Reset()
		}

		if currLine.Len() > 0 {
			currLine.WriteString(" ")
		}

		currLine.WriteString(word)
	}

	result.WriteString(currLine.String())
	return result.String()
}

func (c *Command) fullPath() string {
	if c.parent == nil {
		return c.name
	}
	return c.parent.fullPath() + " " + c.name
}
