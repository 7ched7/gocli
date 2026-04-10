package gocli

import (
	"io"
	"os"
)

// App represents the main CLI application.
// It manages application commands, global flags, configurations, and I/O streams.
type App struct {
	root    *Command
	version string
	config  AppConfig
	stdout  io.Writer
	stderr  io.Writer
}

// AppInfo provides access to application metadata and behaviour.
type AppInfo interface {
	Name() string                       // Name returns the display name of the application.
	Version() string                    // Version returns the version of the application.
	Description() string                // Description returns the description of the application.
	Commands() []CommandInfo            // Commands returns all registered top-level commands.
	GlobalFlags() []FlagInfo            // GlobalFlags returns all registered global flags.
	MinArg() int                        // MinArg returns the minimum number of positional arguments.
	MaxArg() int                        // MaxArg returns the maximum number of positional arguments.
	Config() AppConfig                  // Config returns the configuration settings of the application.
	Stdout() io.Writer                  // Stdout returns the output writer used for standard output.
	Stderr() io.Writer                  // Stderr returns the output writer used for standard error.
	Help() string                       // Help generates and returns the global help menu for the application.
	CommandHelp(cmd CommandInfo) string // CommandHelp generates and returns a help menu for a specific command.
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
	return a.handler(args[1:])
}

// NewApp creates and returns a new App instance
// with the given name and default settings.
func NewApp(name string) *App {
	return &App{
		root: &Command{
			name:        name,
			subcommands: []CommandInfo{},
			flags:       []FlagInfo{},
		},
		config: DefaultAppConfig(),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// WithVersion sets the version for the application.
// This value is displayed when the version flag is used.
func (a *App) WithVersion(version string) *App {
	a.config.VersionFlag = DefaultVersionFlag()
	a.version = version
	return a
}

// WithDescription sets the description for the application.
// The description is shown in application help menu.
func (a *App) WithDescription(description string) *App {
	a.root.long = description
	return a
}

// WithMinArg sets the minimum number of positional arguments required by the application.
func (a *App) WithMinArg(min int) *App {
	a.root.minArg = min
	return a
}

// WithMaxArg sets the maximum number of positional arguments allowed for the application.
func (a *App) WithMaxArg(max int) *App {
	a.root.maxArg = max
	return a
}

// WithConfig sets the configuration settings for the application.
func (a *App) WithConfig(config AppConfig) *App {
	if config.HelpFlag != nil {
		config.HelpFlag.setRole(flagHelp)
	}

	if config.VersionFlag != nil {
		config.VersionFlag.setRole(flagVersion)
	}

	a.config = config
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

// WithAction assigns the default action to be executed when the application is run
// without specifying any command.
func (a *App) WithAction(fn func(ctx *Context) error) *App {
	a.root.actionF = fn
	return a
}

// AddCommand registers top-level commands to the application.
func (a *App) AddCommand(commands ...*Command) *App {
	a.root.AddSubcommand(commands...)
	return a
}

// AddGlobalFlag registers global flags to the application.
// Global flags apply to all commands.
func (a *App) AddGlobalFlag(flags ...FlagInfo) *App {
	a.root.AddFlag(flags...)
	return a
}

// Name returns the display name of the application.
func (a *App) Name() string { return a.root.name }

// Version returns the version of the application.
func (a *App) Version() string { return a.version }

// Description returns the description of the application.
func (a *App) Description() string { return a.root.long }

// Commands returns all registered top-level commands.
func (a *App) Commands() []CommandInfo { return a.root.subcommands }

// GlobalFlags returns all registered global flags.
func (a *App) GlobalFlags() []FlagInfo { return a.root.flags }

// MinArg returns the minimum number of positional arguments required by the application.
// If not set, it returns 0.
func (a *App) MinArg() int { return a.root.minArg }

// MaxArg returns the maximum number of positional arguments allowed for the application.
// If not set, it returns 0.
func (a *App) MaxArg() int { return a.root.maxArg }

// Config returns the configuration settings of the application.
func (a *App) Config() AppConfig { return a.config }

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
