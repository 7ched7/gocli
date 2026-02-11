package gocli

import (
	"io"
	"os"
)

// App represents the main CLI application.
// It manages application commands, configurations, and I/O streams.
type App struct {
	name             string                                            // Application name
	version          string                                            // Application version
	description      string                                            // Application description
	commands         []*Command                                        // Registered commands
	stdout           io.Writer                                         // Standard output
	stderr           io.Writer                                         // Standard error
	customMessageMap map[errorType]func(app *App, err CLIError) string // Message map for customized messages
}

// Run starts the application with os.Args.
// It is the main entry point when the CLI is executed.
func (a *App) Run() int {
	return a.RunWithArgs(os.Args)
}

// RunWithArgs starts the application with a custom set of arguments.
// It is useful for testing or integrating the CLI with another program.
func (a *App) RunWithArgs(args []string) int {
	if code := a.handleGlobalArgs(args); code != -1 {
		return code
	}

	cmd, code := a.findRootCommand(args[1])
	if code != -1 {
		return code
	}

	cmd, remainingArgs := a.findSubcommand(cmd, args[2:])

	args, flags, err := a.parseCommand(cmd, remainingArgs)
	if err != -1 {
		return err
	}

	a.runCommand(cmd, args, *flags)
	return 0
}

// NewApp creates and returns a new App instance
// with the given name and default settings.
func NewApp(name string) *App {
	return &App{
		name:             name,
		commands:         []*Command{},
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		customMessageMap: map[errorType]func(app *App, err CLIError) string{},
	}
}

// WithVersion sets the version for the application.
// This value is displayed when the --version or -v flag is used.
func (a *App) WithVersion(version string) *App {
	a.version = version
	return a
}

// WithDescription sets the description for the application.
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

// Name returns the name of the application.
func (a *App) Name() string { return a.name }

// Version returns the version of the application.
func (a *App) Version() string { return a.version }

// Description returns the description of the application.
func (a *App) Description() string { return a.description }

// Commands returns all registered top-level commands.
func (a *App) Commands() []*Command { return a.commands }

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
