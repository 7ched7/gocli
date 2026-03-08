package gocli

import (
	"io"
	"os"
)

// App represents the main CLI application.
// It manages application commands, global flags, configurations, and I/O streams.
type App struct {
	name              string
	version           string
	description       string
	commands          []*Command
	root              *Command
	globalFlags       []FlagInfo
	stdout            io.Writer
	stderr            io.Writer
	customMessagesMap map[messageType]func(msgCtx MessageContext) string
}

const (
	stateContinue = -1
	exitOK        = 0
	exitError     = 1
	exitUsage     = 2
)

// Run starts the application with os.Args.
// It is the main entry point when the CLI is executed.
func (a *App) Run() int {
	return a.RunWithArgs(os.Args)
}

// RunWithArgs starts the application with a custom set of arguments.
// It is useful for testing or integrating the CLI with another program.
func (a *App) RunWithArgs(args []string) int {
	ctx, cmd, code := a.parseCommand(args[1:])
	if code != stateContinue {
		return code
	}

	a.runCommand(ctx, cmd)
	return exitOK
}

// NewApp creates and returns a new App instance
// with the given name and default settings.
func NewApp(name string) *App {
	return &App{
		name:              name,
		commands:          []*Command{},
		root:              &Command{},
		globalFlags:       []FlagInfo{},
		stdout:            os.Stdout,
		stderr:            os.Stderr,
		customMessagesMap: map[messageType]func(msgCtx MessageContext) string{},
	}
}

// WithVersion sets the version for the application.
// This value is displayed when the version flag is used.
func (a *App) WithVersion(version string) *App {
	a.version = version
	return a
}

// WithDescription sets the description for the application.
// The description is shown in application help menu.
func (a *App) WithDescription(description string) *App {
	a.description = description
	return a
}

// WithStdout sets the writer used for standard output.
func (a *App) WithStdout(out io.Writer) *App {
	a.stdout = out
	return a
}

// WithStderr sets the writer used for error output.
func (a *App) WithStderr(err io.Writer) *App {
	a.stderr = err
	return a
}

// Name returns the display name of the application.
func (a *App) Name() string { return a.name }

// Version returns the version of the application.
func (a *App) Version() string { return a.version }

// Description returns the description of the application.
func (a *App) Description() string { return a.description }

// Commands returns all registered top-level commands.
func (a *App) Commands() []*Command { return a.commands }

// GlobalFlags returns all registered global flags.
func (a *App) GlobalFlags() []FlagInfo { return a.globalFlags }

// Stdout returns the output writer used for standard output.
// If nil, it falls back to os.Stdout.
func (a *App) Stdout() io.Writer {
	if a.stdout != nil {
		return a.stdout
	}
	return os.Stdout
}

// Stderr returns the output writer used for standard error.
// If nil, it falls back to os.Stderr.
func (a *App) Stderr() io.Writer {
	if a.stderr != nil {
		return a.stderr
	}
	return os.Stderr
}
