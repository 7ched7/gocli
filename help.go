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

	cmdRows := commandsToRows(a.root.subcommands)
	flagRows := flagsToRows(a.getFlags(a.root))

	// Usage
	a.writeUsage(&sb, a.root)

	// Description
	writeDescription(&sb, a.root.long)

	// Commands
	writeSection(&sb, "Commands", cmdRows)

	// Global flags
	writeSection(&sb, "Global Flags", flagRows)

	// Footer
	if len(a.root.subcommands) > 0 {
		a.writeFooter(&sb)
	}

	return sb.String()
}

// CommandHelp generates and returns a help menu for a specific command.
// It includes the full command path, command description, registered subcommands and flags.
func (a *App) CommandHelp(cmd CommandInfo) string {
	var sb strings.Builder

	cmdRows := commandsToRows(cmd.Subcommands())
	flagRows := flagsToRows(a.getFlags(cmd))

	// Usage
	a.writeUsage(&sb, cmd)

	// Description
	writeDescription(&sb, cmd.Long())

	// Commands
	writeSection(&sb, "Commands", cmdRows)

	// Flags
	writeSection(&sb, "Flags", flagRows)

	return sb.String()
}

func (a *App) getFlags(cmd CommandInfo) []FlagInfo {
	displayFlags := make([]FlagInfo, 0, len(cmd.Flags()))
	displayFlags = append(displayFlags, cmd.Flags()...)

	if a.config.HelpFlag != nil {
		displayFlags = append(displayFlags, a.config.HelpFlag)
	}

	if cmd == a.root && a.config.VersionFlag != nil {
		displayFlags = append(displayFlags, a.config.VersionFlag)
	}

	return displayFlags
}

func commandsToRows(cmds []CommandInfo) []row {
	rows := make([]row, 0)

	for _, c := range cmds {
		if c.Name() == "" && c.Alias() == "" {
			continue
		}

		left := ""
		if name := c.Name(); name != "" {
			left += name
		}

		if c.Name() != "" && c.Alias() != "" {
			left += ", "
		}

		if alias := c.Alias(); alias != "" {
			left += alias
		}

		rows = append(rows, row{left, c.Short(), len(left)})
	}

	return rows
}

func flagsToRows(flags []FlagInfo) []row {
	rows := make([]row, 0)

	for _, f := range flags {
		if f.Name() == "" && f.Alias() == "" {
			continue
		}

		left := ""
		if alias := f.Alias(); alias != "" {
			left = "-" + alias

			if f.Name() != "" {
				left += ", "
			}
		} else {
			left = "  "

			if f.Name() != "" {
				left += "  "
			}
		}

		if name := f.Name(); name != "" {
			left += "--" + f.Name()
		}

		switch f.Value().Get().(type) {
		case bool:
		default:
			if f.Metavar() != "" {
				left += " " + f.Metavar()
			}
		}

		rows = append(rows, row{left, f.Description(), len(left)})
	}

	return rows
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
	return fmt.Sprintf("\n  %-*s  %s", maxKeyLen, left, wrap(right, maxKeyLen+4, leftExceeds))
}

func (a *App) writeUsage(sb *strings.Builder, cmd CommandInfo) {
	isRoot := cmd == a.root
	hasCommand := (isRoot && len(cmd.Subcommands()) > 0) || (!isRoot && len(cmd.Subcommands()) > 0)
	hasCmdFlag := !isRoot && len(cmd.Flags()) > 0
	hasGlobalFlag := len(a.root.flags) > 0
	hasArg := !(cmd.MinArg() == 0 && cmd.MaxArg() == 0)

	writeBase := func() {
		sb.WriteString("  " + a.root.name)

		if hasGlobalFlag {
			sb.WriteString(" [global flags]")
		}

		sb.WriteString(fullPath(cmd))

		if hasCmdFlag {
			sb.WriteString(" [flags]")
		}
	}

	writeArgs := func() {
		if cmd.MinArg() == 0 {
			if cmd.MaxArg() == 1 {
				sb.WriteString(" [arg]")
			} else {
				sb.WriteString(" [arg]...")
			}
		} else {
			if cmd.MaxArg() == 1 {
				sb.WriteString(" <arg>")
			} else {
				sb.WriteString(" <arg>...")
			}
		}
	}

	sb.WriteString("Usage:\n")
	writeBase()

	if hasCommand {
		if cmd.action() == nil && cmd.MinArg() == 0 && cmd.MaxArg() == 0 {
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
}

func writeDescription(sb *strings.Builder, text string) {
	if text == "" {
		return
	}
	sb.WriteString("\n\n" + wrap(text, 0, false))
}

func writeSection(sb *strings.Builder, title string, rows []row) {
	if len(rows) == 0 {
		return
	}
	maxKeyLen := getMaxKeyLen(rows)
	sb.WriteString("\n\n" + title + ":")
	for _, r := range rows {
		leftExceeds := r.leftWidth > maxKeyWidth
		sb.WriteString(writeRow(r.left, r.right, leftExceeds, maxKeyLen))
	}
}

func (a *App) writeFooter(sb *strings.Builder) {
	var footer string
	if a.config.HelpFlag != nil {
		h := flagDisplayName(a.config.HelpFlag, true)
		if h != "" {
			footer = fmt.Sprintf("\n\nUse \"%s <command> %s\" for more information about a command.", a.root.name, h)
		}
	}
	sb.WriteString(footer)
}

func flagDisplayName(f FlagInfo, dash bool) string {
	if f == nil {
		return ""
	}

	name := f.Name()

	if name == "" {
		alias := f.Alias()

		if dash && alias != "" {
			return "-" + alias
		}
		return alias
	}

	if dash {
		return "--" + name
	}
	return name
}

func commandDisplayName(c CommandInfo) string {
	if c == nil {
		return ""
	}

	if c.Name() == "" {
		return c.Alias()
	}
	return c.Name()
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

func fullPath(c CommandInfo) string {
	if c.Parent() == nil {
		return ""
	}
	return fullPath(c.Parent()) + " " + c.Name()
}
