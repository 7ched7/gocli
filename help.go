package gocli

import (
	"fmt"
	"strings"
)

type row struct {
	left      string
	right     string
	leftWidth int
}

const tWidth = 80
const maxKeyWidth = 25

// Help generates and returns the global help menu for the application.
// It includes usage instructions, application description, registered commands and global flags.
func (a *App) Help() string {
	var sb strings.Builder

	cmdRows := commandsToRows(a.commands)
	flagRows := flagsToRows(a.globalFlags)

	flagRows = append(flagRows,
		row{left: "-h, --help", right: "Show help", leftWidth: 10},
		row{left: "    --version", right: "Show version", leftWidth: 13},
	)

	// Usage
	a.writeUsage(&sb, a.root)

	// Description
	writeDescription(&sb, a.description)

	// Commands
	writeSection(&sb, "Commands", cmdRows)

	// Global flags
	writeSection(&sb, "Global Flags", flagRows)

	// Footer
	if len(a.commands) > 0 {
		sb.WriteString("\nUse \"" + a.name + " <command> --help\" for more information about a command.\n")
	}

	return sb.String()
}

// CommandHelp generates and returns a help menu for a specific command.
// It includes the full command path, command description, registered subcommands and flags.
func (a *App) CommandHelp(cmd *Command) string {
	var sb strings.Builder

	cmdRows := commandsToRows(cmd.subcommands)
	flagRows := flagsToRows(cmd.flags)

	// Usage
	a.writeUsage(&sb, cmd)

	// Description
	writeDescription(&sb, cmd.long)

	// Commands
	writeSection(&sb, "Commands", cmdRows)

	// Flags
	writeSection(&sb, "Flags", flagRows)

	return sb.String()
}

func wrap(text string, indent int, leftExceeds bool) string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	var currLine strings.Builder

	if leftExceeds {
		result.WriteString("\n" + strings.Repeat(" ", indent))
	}

	for _, word := range words {
		// Start a new line if it exceeds width
		if currLine.Len()+len(word)+1 > tWidth-indent {
			result.WriteString(currLine.String() + "\n" + strings.Repeat(" ", indent))
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

func writeRow(left, right string, leftExceeds bool, maxKeyLen int) string {
	return fmt.Sprintf("  %-*s  %s\n", maxKeyLen, left, wrap(right, maxKeyLen+4, leftExceeds))
}

func commandsToRows(cmds []*Command) []row {
	rows := make([]row, 0)

	for _, c := range cmds {
		left := c.name
		if c.alias != "" {
			left += ", " + c.alias
		}
		rows = append(rows, row{left, c.short, len(left)})
	}

	return rows
}

func flagsToRows(flags []FlagInfo) []row {
	rows := make([]row, 0)

	for _, f := range flags {
		left := ""
		if alias := f.Alias(); alias != "" {
			left = fmt.Sprintf("-%s, ", alias)
		} else {
			left = "    "
		}

		left += "--" + f.Name()

		switch f.Value().Get().(type) {
		case bool:
		default:
			if f.Metavar() != "" {
				left += " " + f.Metavar()
			}
		}

		right := f.Description()

		rows = append(rows, row{left, right, len(left)})
	}

	return rows
}

func getMaxKeyLen(rows []row) int {
	max := 0
	for _, r := range rows {
		if r.leftWidth > max {
			max = r.leftWidth
		}
	}
	if max > maxKeyWidth {
		max = maxKeyWidth
	}
	return max
}

func (a *App) writeUsage(sb *strings.Builder, cmd *Command) {
	isRoot := cmd == a.root
	hasCommand := (isRoot && len(a.commands) > 0) || (!isRoot && len(cmd.subcommands) > 0)
	hasCmdFlag := !isRoot && len(cmd.flags) > 0
	hasGlobalFlag := len(a.globalFlags) > 0
	hasArg := !(cmd.minArg == 0 && cmd.maxArg == 0)

	writeBase := func() {
		sb.WriteString("  " + a.name)

		if hasGlobalFlag {
			sb.WriteString(" [global flags]")
		}

		sb.WriteString(cmd.fullPath())

		if hasCmdFlag {
			sb.WriteString(" [flags]")
		}
	}

	writeArgs := func() {
		if cmd.minArg == 0 {
			if cmd.maxArg == 1 {
				sb.WriteString(" [arg]")
			} else {
				sb.WriteString(" [arg]...")
			}
		} else {
			if cmd.maxArg == 1 {
				sb.WriteString(" <arg>")
			} else {
				sb.WriteString(" <arg>...")
			}
		}
	}

	sb.WriteString("Usage:\n")
	writeBase()

	if hasCommand {
		if cmd.action == nil && cmd.minArg == 0 && cmd.maxArg == 0 {
			sb.WriteString(" <command>")
		} else {
			sb.WriteString(" [command]")
		}
	}

	if hasArg && hasCommand {
		sb.WriteString("\n")
		writeBase()
		writeArgs()
	} else if hasArg {
		writeArgs()
	}

	sb.WriteString("\n")
}

func writeSection(sb *strings.Builder, title string, rows []row) {
	if len(rows) == 0 {
		return
	}
	maxKeyLen := getMaxKeyLen(rows)
	sb.WriteString("\n" + title + ":\n")
	for _, r := range rows {
		leftExceeds := r.leftWidth > maxKeyWidth
		sb.WriteString(writeRow(r.left, r.right, leftExceeds, maxKeyLen))
	}
}

func writeDescription(sb *strings.Builder, text string) {
	if text == "" {
		return
	}
	sb.WriteString("\n" + wrap(text, 0, false) + "\n")
}

func (c *Command) fullPath() string {
	if c.parent == nil {
		return c.name
	}
	return c.parent.fullPath() + " " + c.name
}
