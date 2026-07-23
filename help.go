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
func (a *App) Help() string {
	commandRows := commandsToRows(a.root.subcommands)
	globalFlagRows := flagsToRows(a.root.flags)
	defaultFlagRows := flagsToRows(a.defaultFlags(true))
	globalFlagRows = append(globalFlagRows, defaultFlagRows...)

	var sb strings.Builder

	sb.WriteString(renderUsage(a.root))
	sb.WriteString(renderDescription(a.root.long))
	sb.WriteString(renderSection("Commands", commandRows))
	sb.WriteString(renderSection("Global Flags", globalFlagRows))

	if len(a.root.subcommands) > 0 {
		sb.WriteString(renderFooter(a))
	}

	return sb.String()
}

// Help generates and returns the help menu for the command.
func (c *Command) Help() string {
	commandRows := commandsToRows(c.subcommands)
	flagRows := flagsToRows(c.flags)
	globalFlagRows := flagsToRows(c.app.root.flags)
	defaultFlagRows := flagsToRows(c.app.defaultFlags(false))
	globalFlagRows = append(globalFlagRows, defaultFlagRows...)

	var sb strings.Builder

	sb.WriteString(renderUsage(c))
	sb.WriteString(renderDescription(c.long))
	sb.WriteString(renderSection("Commands", commandRows))
	sb.WriteString(renderSection("Flags", flagRows))
	sb.WriteString(renderSection("Global Flags", globalFlagRows))

	return sb.String()
}

// Path returns the execution path of the command.
func (c *Command) Path() string {
	if c.parent == nil {
		return c.name
	}
	return c.parent.Path() + " " + c.name
}

// Usage returns the usage information of the command.
func (c *Command) Usage() string {
	isRoot := c.parent == nil
	hasArgs := len(c.arguments) > 0
	hasGlobalFlags := len(c.app.root.flags) > 0
	hasCommandFlags := !isRoot && len(c.flags) > 0
	hasCommands := len(c.subcommands) > 0

	base := func() string {
		p := c.Path()
		parts := strings.Fields(p)

		if len(parts) == 0 {
			return p
		}

		result := []string{parts[0]}
		if hasGlobalFlags {
			result = append(result, "[global flags]")
		}
		result = append(result, parts[1:]...)
		if hasCommandFlags {
			result = append(result, "[flags]")
		}

		return strings.Join(result, " ")
	}

	args := func() string {
		var out strings.Builder

		for _, a := range c.arguments {
			if a.Min() == 0 {
				if a.Max() == 1 {
					out.WriteString(" [")
					out.WriteString(a.Name())
					out.WriteString("]")
				} else {
					out.WriteString(" [")
					out.WriteString(a.Name())
					out.WriteString("]...")
				}
			} else {
				if a.Max() == 1 {
					out.WriteString(" <")
					out.WriteString(a.Name())
					out.WriteString(">")
				} else {
					out.WriteString(" <")
					out.WriteString(a.Name())
					out.WriteString(">...")
				}
			}
		}

		return out.String()
	}

	out := base()

	if hasCommands {
		if c.action == nil && !hasArgs {
			out += " <command>"
		} else {
			out += " [command]"
		}
	}

	if hasArgs && hasCommands {
		out += "\n"
		out += base()
		out += args()
	} else if hasArgs {
		out += args()
	}

	return out
}

func (a *App) defaultFlags(includeVersion bool) []FlagInfo {
	df := make([]FlagInfo, 0, 2)

	if a.Config().HelpFlag != nil {
		df = append(df, a.Config().HelpFlag)
	}
	if includeVersion && a.Config().VersionFlag != nil {
		df = append(df, a.Config().VersionFlag)
	}
	return df
}

func commandsToRows(commands []CommandInfo) []row {
	rows := make([]row, 0)

	for _, c := range commands {
		name := c.Name()
		alias := c.Alias()

		left := name

		if alias != "" {
			left += ", " + alias
		}

		rows = append(rows, row{left, c.Short(), len(left)})
	}

	return rows
}

func flagsToRows(flags []FlagInfo) []row {
	rows := make([]row, 0)

	for _, f := range flags {
		name := f.Name()
		shorthand := f.Shorthand()
		placeholder := f.Placeholder()

		left := ""

		if shorthand != "" {
			left = "-" + shorthand

			if name != "" {
				left += ", "
			}
		} else {
			left = "  "

			if name != "" {
				left += "  "
			}
		}

		if name != "" {
			left += "--" + name
		}

		if placeholder != "" {
			if name != "" {
				left += "="
			} else {
				left += " "
			}
			left += placeholder
		}

		rows = append(rows, row{left, f.Description(), len(left)})
	}

	return rows
}

func renderUsage(c CommandInfo) string {
	lines := strings.Split(c.Usage(), "\n")
	out := "Usage:\n  "

	if len(lines) > 1 {
		out += fmt.Sprintf("%s\n  %s", lines[0], lines[1])
	} else {
		out += fmt.Sprintf("%s", lines[0])
	}
	return out
}

func renderDescription(text string) string {
	if text == "" {
		return ""
	}
	return "\n\n" + wrap(text, 0, false)
}

func renderSection(title string, rows []row) string {
	if len(rows) == 0 {
		return ""
	}

	maxKeyLen := getMaxKeyLen(rows)

	var out strings.Builder
	out.WriteString("\n\n")
	out.WriteString(title)
	out.WriteString(":")

	for _, r := range rows {
		out.WriteString(renderRow(
			r.left,
			r.right,
			r.leftWidth > maxKeyWidth,
			maxKeyLen,
		))
	}

	return out.String()
}

func renderFooter(a AppInfo) string {
	if a.Config().HelpFlag == nil {
		return ""
	}

	name := flagDisplayName(a.Config().HelpFlag, true)
	return fmt.Sprintf(
		"\n\nUse \"%s <command> %s\" for more information about a command.",
		a.Name(),
		name,
	)
}

func renderRow(left, right string, leftExceeds bool, maxKeyLen int) string {
	return fmt.Sprintf("\n  %-*s  %s", maxKeyLen, left, wrap(right, maxKeyLen+4, leftExceeds))
}

func wrap(text string, indent int, leftExceeds bool) string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	var currLine strings.Builder

	if leftExceeds {
		result.WriteString("\n")
		result.WriteString(strings.Repeat(" ", indent))
	}

	for _, w := range words {
		// Start a new line if it exceeds width
		if currLine.Len()+len(w)+1 > tWidth-indent {
			result.WriteString(currLine.String())
			result.WriteString("\n")
			result.WriteString(strings.Repeat(" ", indent))
			currLine.Reset()
		}

		if currLine.Len() > 0 {
			currLine.WriteString(" ")
		}

		currLine.WriteString(w)
	}

	result.WriteString(currLine.String())
	return result.String()
}

func flagDisplayName(f FlagInfo, includeDash bool) string {
	name := f.Name()

	if name == "" {
		shorthand := f.Shorthand()

		if includeDash {
			return "-" + shorthand
		}
		return shorthand
	}

	if includeDash {
		return "--" + name
	}
	return name
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
