package gocli

import (
	"errors"
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
	Name() string              // Name returns the display name of the application.
	Version() string           // Version returns the version of the application.
	Description() string       // Description returns the description of the application.
	Commands() []CommandInfo   // Commands returns all registered top-level commands.
	Arguments() []ArgumentInfo // Arguments returns all registered arguments.
	GlobalFlags() []FlagInfo   // GlobalFlags returns all registered global flags.
	Config() AppConfig         // Config returns the configuration settings of the application.
	Help() string              // Help generates and returns the global help menu for the application.
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
		return exitError
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

// Verify traverses the application tree, verifies components,
// and returns a combined error if any occurs.
func (a *App) Verify() error {
	errs := a.verify()
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("verification failed with %d issues:\n%v",
		len(errs),
		errors.Join(errs...),
	)
}

// NewApp creates and returns a new App instance with the given name.
func NewApp(name string) *App {
	a := &App{
		root:   NewCommand(name),
		config: DefaultAppConfig(),
	}
	a.root.setApp(a)
	return a
}

// WithVersion sets the version for the application.
// This value is displayed when the version flag is used.
func (a *App) WithVersion(version string) *App {
	if a.config.VersionFlag == nil {
		a.config.VersionFlag = DefaultVersionFlag()
	}
	a.version = version
	return a
}

// WithDescription sets the description for the application.
// The description is shown in help menu.
func (a *App) WithDescription(description string) *App {
	a.root.long = description
	return a
}

// WithConfig sets the configuration settings for the application.
func (a *App) WithConfig(config AppConfig) *App {
	a.config = config
	return a
}

// WithAction assigns the default action to be executed when the application is run
// without specifying any command.
func (a *App) WithAction(fn func(ctx *Context) error) *App {
	a.root.action = fn
	return a
}

// AddCommand registers top-level commands to the application.
func (a *App) AddCommand(commands ...CommandInfo) *App {
	a.root.AddSubcommand(commands...)
	a.bindCommands(a.root, commands...)
	return a
}

func (a *App) bindCommands(p CommandInfo, commands ...CommandInfo) {
	for _, c := range commands {
		c.setParent(p)
		c.setApp(a)
		a.bindCommands(c, c.Subcommands()...)
	}
}

// AddArgument registers arguments to the application.
func (a *App) AddArgument(arguments ...ArgumentInfo) *App {
	a.root.AddArgument(arguments...)
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

// Arguments returns all registered arguments.
func (a *App) Arguments() []ArgumentInfo { return a.root.arguments }

// GlobalFlags returns all registered global flags.
func (a *App) GlobalFlags() []FlagInfo { return a.root.flags }

// Config returns the configuration settings of the application.
func (a *App) Config() AppConfig { return a.config }
