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

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s [COMMAND] [FLAGS] [ARGS...]\n", a.name)
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "Commands:")

	for _, cmd := range a.commands {
		entry := cmd.name

		if cmd.alias != "" {
			entry += ", " + cmd.alias
		}

		fmt.Fprintf(&sb, "  %-18s %s\n", entry, cmd.short)
	}

	fmt.Fprintln(&sb, "\nFlags:")
	fmt.Fprintf(&sb, "  %-18s %s\n", "--help, -h", "Show help")
	if a.version != "" {
		fmt.Fprintf(&sb, "  %-18s %s\n", "--version, -v", "Show version")
	}

	if len(a.commands) > 0 {
		fmt.Fprintf(&sb, "\nFor more information about a command, use '%s <command> --help'.\n", a.name)
	}

	return sb.String()
}

// CommandHelp generates and returns a help menu for a specific command.
// It includes the full command path, argument expectations, registered subcommands/flags,
// and is displayed when the user types <cmd> --help or <cmd> -h.
func (a *App) CommandHelp(cmd *Command) string {
	var sb strings.Builder

	fmt.Fprintln(&sb, "Usage:")
	fmt.Fprintf(&sb, "  %s", a.name)

	// Build full command path
	var parents []string
	currCmd := cmd
	for currCmd.Parent() != nil {
		currCmd = currCmd.Parent()
		parents = append(parents, currCmd.name)
	}

	for i := len(parents) - 1; i >= 0; i-- {
		fmt.Fprintf(&sb, " %s", parents[i])
	}

	fmt.Fprintf(&sb, " %s", cmd.name)

	hasSubcmd := len(cmd.subcommands) > 0
	hasFlag := len(cmd.flags) > 0

	if hasSubcmd {
		fmt.Fprint(&sb, " [COMMAND]")
	}
	if hasFlag {
		fmt.Fprint(&sb, " [FLAGS]")
	}

	if cmd.maxArg == 1 {
		fmt.Fprint(&sb, " [ARG]")
	} else if cmd.maxArg > 1 {
		fmt.Fprintf(&sb, " [ARG1...ARG%d]", cmd.maxArg)
	} else if cmd.minArg > 0 {
		fmt.Fprint(&sb, " [ARGS...]")
	}

	fmt.Fprintln(&sb)
	if cmd.long != "" {
		fmt.Fprintf(&sb, "\n%s\n", cmd.long)
	}

	if hasSubcmd {
		fmt.Fprintln(&sb, "\nCommands:")
		for _, f := range cmd.subcommands {
			displayName := f.name
			if f.alias != "" {
				displayName += ", " + f.alias
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.short)
		}
	}

	if hasFlag {
		fmt.Fprintln(&sb, "\nFlags:")
		for _, f := range cmd.flags {
			displayName := "--" + f.Name()
			if f.Alias() != "" {
				displayName += ", -" + f.Alias()
			}
			fmt.Fprintf(&sb, "  %-18s %s\n", displayName, f.Description())
		}
	}

	return sb.String()
}
