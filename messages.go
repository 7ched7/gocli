package gocli

import (
	"errors"
	"fmt"
	"io"
)

type messageType int

const (
	msgNone messageType = iota
	MsgHelp
	MsgCommandHelp
	MsgVersion
	MsgNoCommand
	MsgUnknownCommand
	MsgSubcommandRequired
	MsgInvalidFlag
	MsgFlagValueMissing
	MsgFlagRequired
	MsgUnexpectedArgument
	MsgTooFewArguments
	MsgTooManyArguments
	MsgUnsupportedFlagType
)

// CLIMessage represents a message object used by the CLI.
// It includes an exit code, message, message type, optional command pointer,
// error, writer, and extra metadata.
type CLIMessage struct {
	code        int
	message     string
	messageType messageType
	command     CommandInfo
	err         error
	writer      io.Writer
	data        map[string]string
}

// Exit creates a new CLI message with the provided message and code.
func Exit(message string, code int) *CLIMessage {
	return &CLIMessage{message: message, code: code}
}

// WithWriter sets the writer for the message.
func (m *CLIMessage) WithWriter(writer io.Writer) *CLIMessage {
	m.writer = writer
	return m
}

// Error implements the error interface for the message.
// It returns the underlying error message string.
func (m *CLIMessage) Error() string {
	return m.message
}

// Code returns the exit code associated with the event.
func (m *CLIMessage) Code() int { return m.code }

// Message returns the message of the event.
func (m *CLIMessage) Message() string { return m.message }

// MessageType returns the categorized type of the event.
func (m *CLIMessage) MessageType() messageType { return m.messageType }

// Command returns the command where the event occured.
func (m *CLIMessage) Command() CommandInfo { return m.command }

// Data returns a map of metadata related to the event.
func (m *CLIMessage) Data() map[string]string { return m.data }

// Writer returns the writer of the message.
func (m *CLIMessage) Writer() io.Writer {
	return m.writer
}

// MessageContext holds the application instance
// and message object providing details about the event.
type MessageContext struct {
	app AppInfo
	msg *CLIMessage
}

// App returns the application instance.
func (m *MessageContext) App() AppInfo { return m.app }

// Msg returns the message object providing details about the event.
func (m *MessageContext) Msg() *CLIMessage { return m.msg }

func (a *App) getMessageAndExitCode(messageType messageType, cmd CommandInfo, data map[string]string) (string, int) {
	name := func() string {
		if cmd == nil || cmd == a.root {
			return a.root.name
		}
		return cmd.Name()
	}

	var usageMsg string
	if a.config.HelpFlag != nil {
		h := a.config.HelpFlag.Name()
		if h != "" {
			usageMsg = fmt.Sprintf("use --%s for usage information.\n", h)
		}
	}

	switch messageType {
	case MsgHelp:
		return a.Help(), exitOK

	case MsgCommandHelp:
		return a.CommandHelp(cmd), exitOK

	case MsgVersion:
		return fmt.Sprintf("%s version %s\n", a.root.name, a.version), exitOK

	case MsgNoCommand:
		return a.Help(), exitUsage

	case MsgUnknownCommand:
		return fmt.Sprintf("error: unknown command: '%s'\n%s", data["command"], usageMsg), exitUsage

	case MsgSubcommandRequired:
		return fmt.Sprintf("error: a subcommand is required for the command: '%s'\n%s", data["command"], usageMsg), exitUsage

	case MsgInvalidFlag:
		return fmt.Sprintf("error: invalid flag: '%s'\n%s", data["flag"], usageMsg), exitUsage

	case MsgFlagValueMissing:
		return fmt.Sprintf("error: a value is required for the flag: '%s'\n", data["flag"]), exitUsage

	case MsgFlagRequired:
		return fmt.Sprintf("error: flag is required: '%s'\n", data["flag"]), exitUsage

	case MsgUnexpectedArgument:
		return fmt.Sprintf("error: unexpected argument: '%s'\n'%s' does not accept arguments.\n", data["argument"], name()), exitUsage

	case MsgTooFewArguments:
		return fmt.Sprintf("error: '%s' requires at least %d argument(s), but got %s.\n", name(), cmd.MinArg(), data["number"]), exitUsage

	case MsgTooManyArguments:
		return fmt.Sprintf("error: '%s' accepts at most %d argument(s), but got %s.\n", name(), cmd.MaxArg(), data["number"]), exitUsage

	case MsgUnsupportedFlagType:
		return fmt.Sprintf("internal error: unsupported flag type: '%s'\n", data["value"]), exitError

	default:
		return "internal error: an unexpected error occurred.\n", exitError
	}
}

func (a *App) exit(message *CLIMessage) int {
	var msg string
	var err = message.err
	code := message.code
	out := a.Stderr()

	if message.messageType != msgNone {
		m, c := a.getMessageAndExitCode(
			message.messageType,
			message.command,
			message.data,
		)

		if code == 0 {
			code = c
		}

		msg = m

		// override the default message
		if a.config.Messages != nil {
			if fn, ok := a.config.Messages[message.messageType]; ok {
				err = fn(MessageContext{
					app: a,
					msg: &CLIMessage{
						code:        code,
						messageType: message.messageType,
						command:     message.command,
						message:     msg,
						data:        message.data,
					},
				})
			}
		}
	}

	if err != nil {
		msg = err.Error()
	}

	if code == exitOK {
		out = a.Stdout()
	}

	var cm *CLIMessage
	if errors.As(err, &cm) {
		code = cm.code

		if cm.writer != nil {
			out = cm.writer
		} else if code == exitOK {
			out = a.Stdout()
		}
	}

	fmt.Fprint(out, msg)
	return code
}

func (a *App) cliExit(messageType messageType, command CommandInfo, data map[string]string) int {
	return a.exit(&CLIMessage{
		messageType: messageType,
		command:     command,
		data:        data,
	})
}

func (a *App) appExit(err error, code int) int {
	return a.exit(&CLIMessage{
		err:  err,
		code: code,
	})
}
