package gocli

import "fmt"

type errorType int

const (
	ErrHelp errorType = iota
	ErrCommandHelp
	ErrVersion
	ErrNoCommand
	ErrUnknownCommand
	ErrSubcommandRequired
	ErrInvalidFlag
	ErrFlagValueMissing
	ErrUnexpectedArgument
	ErrTooFewArguments
	ErrTooManyArguments
	ErrInvalidIntValue
	ErrInvalidFloatValue
	ErrInvalidBoolValue
	ErrUnsupportedFlagType
)

// CLIError represents an error object used by the CLI.
// It includes an exit code, message, error type, and optional command pointer,
// as well as extra metadata.
type CLIError struct {
	Code    int               // Exit code
	Message string            // Error message
	Type    errorType         // Error type
	Cmd     *Command          // Command pointer
	Data    map[string]string // Extra information
}

// Error implements the built-in error interface,
// and returns a defined message for the error.
func (e CLIError) Error() string {
	return e.Message
}

var defaultMessageMap = map[errorType]func(*App, *Command, map[string]string) (int, string){
	ErrHelp: func(a *App, _ *Command, _ map[string]string) (int, string) {
		return ExitOK, a.Help()
	},

	ErrCommandHelp: func(a *App, cmd *Command, _ map[string]string) (int, string) {
		return ExitOK, a.CommandHelp(cmd)
	},

	ErrVersion: func(a *App, _ *Command, _ map[string]string) (int, string) {
		return ExitOK, fmt.Sprintf("%s version %s\n", a.name, a.version)
	},

	ErrNoCommand: func(a *App, _ *Command, _ map[string]string) (int, string) {
		return ExitUsage, a.Help()
	},

	ErrUnknownCommand: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("unknown command: '%s'\n", data["command"])
	},

	ErrSubcommandRequired: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("'%s' requires a subcommand\n", data["command"])
	},

	ErrInvalidFlag: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("invalid flag: '%s'\n", data["flag"])
	},

	ErrFlagValueMissing: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("value required for flag: '%s'\n", data["flag"])
	},

	ErrUnexpectedArgument: func(_ *App, cmd *Command, _ map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("'%s' does not accept argument(s)\n", cmd.name)
	},

	ErrTooFewArguments: func(_ *App, cmd *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("'%s' requires at least %d argument(s), got %s\n", cmd.name, cmd.minArg, data["number"])
	},

	ErrTooManyArguments: func(_ *App, cmd *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("'%s' requires at most %d argument(s), got %s\n", cmd.name, cmd.maxArg, data["number"])
	},

	ErrInvalidIntValue: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("int parse error: '%s'\n", data["value"])
	},

	ErrInvalidFloatValue: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("float parse error: '%s'\n", data["value"])
	},

	ErrInvalidBoolValue: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("bool parse error: '%s'\n", data["value"])
	},

	ErrUnsupportedFlagType: func(_ *App, _ *Command, data map[string]string) (int, string) {
		return ExitUsage, fmt.Sprintf("unsupported type: '%s'\n", data["value"])
	},
}

func (a *App) getMessageAndExitCode(errType errorType, cmd *Command, data map[string]string) (int, string) {
	if fn, ok := defaultMessageMap[errType]; ok {
		return fn(a, cmd, data)
	}
	return ExitError, "unknown error"
}

func (a *App) stop(errType errorType, cmd *Command, data map[string]string) int {
	code, message := a.getMessageAndExitCode(errType, cmd, data)

	// create the appropriate CLI error
	cliErr := CLIError{
		Code:    code,
		Type:    errType,
		Cmd:     cmd,
		Message: message,
		Data:    data,
	}

	return a.handleError(cliErr)
}

func (a *App) handleError(err CLIError) int {
	var msg string
	out := a.Stderr()

	if a.customMessageMap != nil {
		if fn, ok := a.customMessageMap[err.Type]; ok {
			msg = fn(a, err) // override the default message
		}
	}

	if msg == "" {
		msg = err.Message
	}

	if err.Code == ExitOK {
		out = a.Stdout()
	}

	fmt.Fprint(out, msg)
	return err.Code
}

// HandleMessage registers a custom message handler for a specific error type.
// The handler function runs whenever an error of the given type occurs.
func (a *App) HandleMessage(errType errorType, fn func(app *App, err CLIError) string) *App {
	a.customMessageMap[errType] = fn
	return a
}
