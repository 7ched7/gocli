package gocli

import (
	"fmt"
	"os"
)

// App represents the main CLI application.
// It holds the root command, version information,
// and global configuration.
type App struct {
	root    *Command
	version string
	config  AppConfig
}

// AppInfo provides access to application metadata.
type AppInfo interface {
	Name() string                       // Name returns the display name of the application.
	Version() string                    // Version returns the version of the application.
	Description() string                // Description returns the description of the application.
	Commands() []CommandInfo            // Commands returns all registered top-level commands.
	GlobalFlags() []FlagInfo            // GlobalFlags returns all registered global flags.
	MinArg() int                        // MinArg returns the minimum number of positional arguments.
	MaxArg() int                        // MaxArg returns the maximum number of positional arguments.
	Config() AppConfig                  // Config returns the configuration settings of the application.
	Help() string                       // Help generates and returns the global help menu for the application.
	CommandHelp(cmd CommandInfo) string // CommandHelp generates and returns a help menu for a specific command.
}

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// Run starts the application with os.Args and returns an exit code.
func (a *App) Run() int {
	err := a.RunWithArgs(os.Args)
	if err == nil {
		return exitOK
	}

	switch e := err.(type) {
	case *CLIMessage:
		fmt.Fprintln(e.writer, e)
		return e.code
	default:
		fmt.Fprintf(a.config.Stderr, "error: %v\n", err)
		return 1
	}
}

// RunE starts the application with os.Args and returns an error if any occurs.
func (a *App) RunE() error {
	return a.RunWithArgs(os.Args)
}

// RunWithArgs starts the application with a custom set of arguments.
// It is useful for testing or integrating the CLI with another program.
func (a *App) RunWithArgs(args []string) error {
	if len(args) == 0 {
		return a.handler([]string{})
	}
	return a.handler(args[1:])
}

// NewApp creates and returns a new App instance with the given name.
func NewApp(name string) *App {
	return &App{
		root: &Command{
			name:        name,
			subcommands: []CommandInfo{},
			flags:       []FlagInfo{},
		},
		config: DefaultAppConfig(),
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
// The description is shown in help menu.
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
		a.config.HelpFlag = config.HelpFlag
	}

	if config.VersionFlag != nil {
		config.VersionFlag.setRole(flagVersion)
		a.config.VersionFlag = config.VersionFlag
	}

	if config.CustomMessages != nil {
		a.config.CustomMessages = config.CustomMessages
	}

	if config.Stdout != nil {
		a.config.Stdout = config.Stdout
	}

	if config.Stderr != nil {
		a.config.Stderr = config.Stderr
	}

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
