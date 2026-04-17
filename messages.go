package gocli

import (
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
	MsgIntParseError
	MsgFloat64ParseError
	MsgBoolParseError
	MsgUnexpectedArgument
	MsgTooFewArguments
	MsgTooManyArguments
)

// CLIMessage represents a structured message used by the CLI.
// It includes an exit code, message, message type, optional command pointer,
// metadata, and I/O writer.
type CLIMessage struct {
	code        int
	message     string
	messageType messageType
	command     CommandInfo
	data        map[string]string
	writer      io.Writer
}

// Error implements the error interface for the message.
// It returns the underlying error message string.
func (m *CLIMessage) Error() string {
	return m.message
}

// Code returns the exit code associated with the message.
func (m *CLIMessage) Code() int { return m.code }

// Message returns the raw text content of the message.
func (m *CLIMessage) Message() string { return m.message }

// MessageType returns the internal type of the message.
func (m *CLIMessage) MessageType() messageType { return m.messageType }

// Command returns the command that triggered the message.
func (m *CLIMessage) Command() CommandInfo { return m.command }

// Data returns the metadata associated with the message.
func (m *CLIMessage) Data() map[string]string { return m.data }

// Writer returns the writer where the message is written.
func (m *CLIMessage) Writer() io.Writer { return m.writer }

func newCLIMessage(
	code int,
	message string,
	messageType messageType,
	command CommandInfo,
	data map[string]string,
	writer io.Writer,
) *CLIMessage {
	return &CLIMessage{
		code:        code,
		message:     message,
		messageType: messageType,
		command:     command,
		data:        data,
		writer:      writer,
	}
}

// Exit creates a new CLI message with the provided code and message.
func Exit(code int, message string) *CLIMessage {
	return newCLIMessage(code, message, msgNone, nil, nil, nil)
}

// Exitf creates a new CLI message with the code and formatted message.
func Exitf(code int, format string, a ...any) *CLIMessage {
	message := fmt.Sprintf(format, a...)
	return newCLIMessage(code, message, msgNone, nil, nil, nil)
}

// MessageContext provides the necessary environment data
// for formatting and handling CLI messages.
type MessageContext struct {
	app AppInfo
	msg *CLIMessage
}

// App returns the application instance.
func (m *MessageContext) App() AppInfo { return m.app }

// Msg returns the underlying CLIMessage.
func (m *MessageContext) Msg() *CLIMessage { return m.msg }

var defaultMessages MessagesMap = MessagesMap{
	MsgHelp:               msgHelp,
	MsgCommandHelp:        msgCommandHelp,
	MsgVersion:            msgVersion,
	MsgNoCommand:          msgNoCommand,
	MsgUnknownCommand:     msgUnknownCommand,
	MsgSubcommandRequired: msgSubcommandRequired,
	MsgInvalidFlag:        msgInvalidFlag,
	MsgFlagValueMissing:   msgFlagValueMissing,
	MsgFlagRequired:       msgFlagRequired,
	MsgIntParseError:      msgIntParseError,
	MsgFloat64ParseError:  msgFloat64ParseError,
	MsgBoolParseError:     msgBoolParseError,
	MsgUnexpectedArgument: msgUnexpectedArgument,
	MsgTooFewArguments:    msgTooFewArguments,
	MsgTooManyArguments:   msgTooManyArguments,
}

func msgHelp(msgCtx MessageContext) error {
	return Exit(exitOK, msgCtx.app.Help())
}

func msgCommandHelp(msgCtx MessageContext) error {
	return Exit(exitOK, msgCtx.app.CommandHelp(msgCtx.msg.command))
}

func msgVersion(msgCtx MessageContext) error {
	return Exitf(
		exitOK,
		"%s version %s",
		msgCtx.app.Name(),
		msgCtx.app.Version(),
	)
}

func msgNoCommand(msgCtx MessageContext) error {
	return Exit(exitUsage, msgCtx.App().Help())
}

func msgUnknownCommand(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: unknown command: '%s'%s",
		msgCtx.msg.data["command"],
		msgUsage(&msgCtx),
	)
}

func msgSubcommandRequired(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: a subcommand is required for command '%s'%s",
		msgCtx.msg.data["command"],
		msgUsage(&msgCtx),
	)
}

func msgInvalidFlag(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid flag: '%s'%s",
		msgCtx.msg.data["flag"],
		msgUsage(&msgCtx),
	)
}

func msgFlagValueMissing(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: a value is required for flag '%s'",
		msgCtx.msg.data["flag"],
	)
}

func msgFlagRequired(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: flag is required: '%s'",
		msgCtx.msg.data["flag"],
	)
}

func msgIntParseError(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid value '%v': expected integer",
		msgCtx.msg.data["value"],
	)
}

func msgFloat64ParseError(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid value '%v': expected float",
		msgCtx.msg.data["value"],
	)
}

func msgBoolParseError(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid value '%v': expected boolean",
		msgCtx.msg.data["value"],
	)
}

func msgUnexpectedArgument(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: unexpected argument: '%s'",
		msgCtx.msg.data["argument"],
	)
}

func msgTooFewArguments(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: too few arguments; expected at least %d, but got %s",
		msgCtx.msg.command.MinArg(),
		msgCtx.msg.data["number"],
	)
}

func msgTooManyArguments(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: too many arguments; expected at most %d, but got %s",
		msgCtx.msg.command.MaxArg(),
		msgCtx.msg.data["number"],
	)
}

func msgUsage(msgCtx *MessageContext) string {
	helpFlag := msgCtx.app.Config().HelpFlag
	h := flagDisplayName(helpFlag, true)

	if h != "" {
		return fmt.Sprintf("\nuse '%s' for usage information.", h)
	}
	return ""
}

func (a *App) exit(cliMsg *CLIMessage) error {
	getWriter := func(code int) io.Writer {
		if code == exitOK {
			return a.config.Stdout
		}
		return a.config.Stderr
	}

	getMessageInfo := func(err error, currCode int) (string, int) {
		msg := err.Error()
		if e, ok := err.(*CLIMessage); ok {
			return msg, e.code
		}
		return msg, currCode
	}

	code := cliMsg.code
	message := cliMsg.message
	messageType := cliMsg.messageType
	command := cliMsg.command
	data := cliMsg.data
	out := getWriter(code)

	if messageType == msgNone {
		return newCLIMessage(code, message, messageType, command, data, out)
	}

	msgCtx := MessageContext{
		app: a,
		msg: newCLIMessage(code, message, messageType, command, data, out),
	}

	if fn, ok := defaultMessages[messageType]; fn != nil && ok {
		if err := fn(msgCtx); err != nil {
			message, code = getMessageInfo(err, code)
		}
	}

	// override the default message
	if a.config.CustomMessages != nil {
		msgCtx.msg.code = code
		msgCtx.msg.message = message
		msgCtx.msg.writer = getWriter(code)

		if fn, ok := a.config.CustomMessages[messageType]; fn != nil && ok {
			if err := fn(msgCtx); err != nil {
				message, code = getMessageInfo(err, code)
			} else {
				message = ""
			}
		}
	}

	out = getWriter(code)
	return newCLIMessage(code, message, messageType, command, data, out)
}

func (a *App) exitWithMsg(messageType messageType, command CommandInfo, data map[string]string) error {
	return a.exit(newCLIMessage(0, "", messageType, command, data, nil))
}

func (a *App) exitWithErr(err error, defaultCode int) error {
	if e, ok := err.(*CLIMessage); ok {
		defaultCode = e.code
	}
	return a.exit(newCLIMessage(defaultCode, err.Error(), msgNone, nil, nil, nil))
}
