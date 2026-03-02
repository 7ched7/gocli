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
	ErrUnsupportedFlagType
)

// CLIError represents an error object used by the CLI.
// It includes an exit code, message, error type, and optional command pointer,
// as well as extra metadata.
type CLIError struct {
	code      int
	message   string
	errorType errorType
	command   *Command
	data      map[string]string
}

// Error implements the built-in error interface
// and returns the error message.
func (e CLIError) Error() string {
	return e.message
}

// Code returns the exit code associated with the error.
func (e *CLIError) Code() int { return e.code }

// Message returns the message of the error.
func (e *CLIError) Message() string { return e.message }

// ErrorType returns the categorized type of the error.
func (e *CLIError) ErrorType() errorType { return e.errorType }

// Command returns the command where the error occured.
func (e *CLIError) Command() *Command { return e.command }

// Data returns a map of metadata related to the error.
func (e *CLIError) Data() map[string]string { return e.data }

// ErrorContext holds the application instance
// and error object providing details about the error.
type ErrorContext struct {
	app *App
	err *CLIError
}

// App returns the application instance.
func (e *ErrorContext) App() *App { return e.app }

// Err returns the CLIError providing details about the error.
func (e *ErrorContext) Err() *CLIError { return e.err }

func (a *App) getMessageAndExitCode(errType errorType, cmd *Command, data map[string]string) (string, int) {
	name := func() string {
		if cmd == nil || cmd == a.root {
			return a.name
		}
		return cmd.name
	}

	switch errType {
	case ErrHelp:
		return a.Help(), ExitOK

	case ErrCommandHelp:
		return a.CommandHelp(cmd), ExitOK

	case ErrVersion:
		return fmt.Sprintf("%s version %s\n", a.name, a.version), ExitOK

	case ErrNoCommand:
		return a.Help(), ExitUsage

	case ErrUnknownCommand:
		return fmt.Sprintf("error: unknown command: '%s'\nuse --help for usage information.\n", data["command"]), ExitUsage

	case ErrSubcommandRequired:
		return fmt.Sprintf("error: a subcommand is required for the command: '%s'\nuse --help for usage information.\n", data["command"]), ExitUsage

	case ErrInvalidFlag:
		return fmt.Sprintf("error: invalid flag: '%s'\nuse --help for usage information.\n", data["flag"]), ExitUsage

	case ErrFlagValueMissing:
		return fmt.Sprintf("error: a value is required for the flag: '%s'\n", data["flag"]), ExitUsage

	case ErrUnexpectedArgument:
		return fmt.Sprintf("error: unexpected argument: '%s'\n'%s' does not accept arguments.\n", data["argument"], name()), ExitUsage

	case ErrTooFewArguments:
		return fmt.Sprintf("error: '%s' requires at least %d argument(s), but got %s.\n", name(), cmd.minArg, data["number"]), ExitUsage

	case ErrTooManyArguments:
		return fmt.Sprintf("error: '%s' accepts at most %d argument(s), but got %s.\n", name(), cmd.maxArg, data["number"]), ExitUsage

	case ErrUnsupportedFlagType:
		return fmt.Sprintf("internal error: unsupported flag type: '%s'\n", data["value"]), ExitUsage

	default:
		return "internal error: an unexpected error occurred.\n", ExitError
	}
}

func (a *App) stop(errorType errorType, cmd *Command, data map[string]string) int {
	msg, code := a.getMessageAndExitCode(errorType, cmd, data)

	cliErr := &CLIError{
		code:      code,
		errorType: errorType,
		command:   cmd,
		message:   msg,
		data:      data,
	}

	errCtx := ErrorContext{
		app: a,
		err: cliErr,
	}

	return a.handleError(errCtx)
}

func (a *App) handleError(errCtx ErrorContext) int {
	var msg string
	isCustom := false
	out := a.Stderr()

	if a.customMessageMap != nil {
		if fn, ok := a.customMessageMap[errCtx.err.errorType]; ok {
			msg = fn(errCtx) // override the default message
			isCustom = true
		}
	}

	if !isCustom {
		msg = errCtx.err.message
	}

	if errCtx.err.code == ExitOK {
		out = a.Stdout()
	}

	fmt.Fprint(out, msg)
	return errCtx.err.code
}

// HandleMessage registers a custom message handler for a specific error type.
// The handler function runs whenever an error of the given type occurs.
func (a *App) HandleMessage(errType errorType, fn func(errCtx ErrorContext) string) *App {
	a.customMessageMap[errType] = fn
	return a
}
