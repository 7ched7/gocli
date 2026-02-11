package gocli

import "fmt"

// errorType represents the type of CLI error that occurred.
type errorType int

// List of all CLI error types.
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
	Code    int            // Exit code
	Message string         // Error message
	Type    errorType      // Error type
	Cmd     *Command       // Command pointer
	Data    map[string]any // Extra information
}

// Error implements the built-in error interface,
// and returns a defined message for the error.
func (e CLIError) Error() string {
	return e.Message
}

// defaultMessageMap holds the application's exit codes and default error messages.
var defaultMessageMap = map[errorType]func(*App, *Command, map[string]any) (int, string){

	ErrHelp: func(a *App, _ *Command, _ map[string]any) (int, string) {
		return 0, a.Help()
	},

	ErrCommandHelp: func(a *App, cmd *Command, _ map[string]any) (int, string) {
		return 0, a.CommandHelp(cmd)
	},

	ErrVersion: func(a *App, _ *Command, _ map[string]any) (int, string) {
		return 0, fmt.Sprintf("%s version %s\n", a.name, a.version)
	},

	ErrNoCommand: func(a *App, _ *Command, _ map[string]any) (int, string) {
		return 0, a.Help()
	},

	ErrUnknownCommand: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("unknown command: '%s'\n", data["command"])
	},

	ErrSubcommandRequired: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("%s requires a subcommand\n", data["command"])
	},

	ErrInvalidFlag: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("invalid flag: '%s'\n", data["flag"])
	},

	ErrFlagValueMissing: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("value required for flag: '%s'\n", data["flag"])
	},

	ErrUnexpectedArgument: func(_ *App, cmd *Command, _ map[string]any) (int, string) {
		return 2, fmt.Sprintf("%s does not accept argument(s)\n", cmd.name)
	},

	ErrTooFewArguments: func(_ *App, cmd *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("%s requires at least %d argument(s), got %d\n", cmd.name, cmd.minArg, data["number"])
	},

	ErrTooManyArguments: func(_ *App, cmd *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("%s requires at most %d argument(s), got %d\n", cmd.name, cmd.maxArg, data["number"])
	},

	ErrInvalidIntValue: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("int parse error: '%s'\n", data["value"])
	},

	ErrInvalidFloatValue: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("float parse error: '%s'\n", data["value"])
	},

	ErrInvalidBoolValue: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("bool parse error: '%s'\n", data["value"])
	},

	ErrUnsupportedFlagType: func(_ *App, _ *Command, data map[string]any) (int, string) {
		return 2, fmt.Sprintf("unsupported type: '%s'\n", fmt.Sprintf("%T", data["value"]))
	},
}

// getMessageAndExitCode returns the appropriate exit code and message
// from the defaultMessageMap.
func getMessageAndExitCode(a *App, errType errorType, cmd *Command, data map[string]any) (int, string) {
	if fn, ok := defaultMessageMap[errType]; ok {
		return fn(a, cmd, data)
	}
	return 2, "unknown error"
}

// stop is a helper function that creates a CLIError with the appropriate code and message,
// and passes it to the handleError, which returns the exit code.
func (a *App) stop(errType errorType, cmd *Command, data map[string]any) int {
	code, message := getMessageAndExitCode(a, errType, cmd, data)

	return a.handleError(CLIError{
		Code:    code,
		Type:    errType,
		Cmd:     cmd,
		Message: message,
		Data:    data,
	})
}

// handleError processes the given error, prints the appropriate message
// to stdout or stderr, and returns the exit code.
func (a *App) handleError(err CLIError) int {
	var msg string
	out := a.Stderr()

	if a.customMessageMap != nil {
		if fn, ok := a.customMessageMap[err.Type]; ok {
			msg = fn(a, err)
		}
	}

	if msg == "" {
		msg = err.Message
	}

	if err.Code == 0 {
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
