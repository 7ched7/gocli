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
// It includes an exit code, message, message type, and optional command pointer,
// as well as extra metadata.
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

// Exit creates a new CLI message with the provided message and code.
func Exit(code int, message string) *CLIMessage {
	return newCLIMessage(code, message, msgNone, nil, nil, nil)
}

// Exitf creates a new CLI message with a formatted message and code.
func Exitf(code int, format string, a ...any) *CLIMessage {
	message := fmt.Sprintf(format, a...)
	return newCLIMessage(code, message, msgNone, nil, nil, nil)
}

// MessageContext provides the necessary environment data for formatting
// and handling CLI messages.
type MessageContext struct {
	app AppInfo
	msg *CLIMessage
}

// App returns the application instance.
func (m *MessageContext) App() AppInfo { return m.app }

// Msg returns the CLIMessage providing details about the event.
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
		"error: unknown command: '%s'\n%s",
		msgCtx.msg.data["command"],
		getUsage(&msgCtx),
	)
}

func msgSubcommandRequired(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: a subcommand is required for the command: '%s'\n%s",
		msgCtx.msg.data["command"],
		getUsage(&msgCtx),
	)
}

func msgInvalidFlag(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid flag: '%s'\n%s",
		msgCtx.msg.data["flag"],
		getUsage(&msgCtx),
	)
}

func msgFlagValueMissing(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: a value is required for the flag: '%s'",
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
		"error: invalid value '%v': must be an integer.",
		msgCtx.msg.data["value"],
	)
}

func msgFloat64ParseError(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid value '%v': must be a float.",
		msgCtx.msg.data["value"],
	)
}

func msgBoolParseError(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: invalid value '%v': must be a bool.",
		msgCtx.msg.data["value"],
	)
}

func msgUnexpectedArgument(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: unexpected argument: '%s'\n'%s' does not accept arguments.",
		msgCtx.msg.data["argument"],
		msgCtx.msg.command.Name(),
	)
}

func msgTooFewArguments(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: '%s' requires at least %d argument(s), but got %s.",
		msgCtx.msg.command.Name(),
		msgCtx.msg.command.MinArg(),
		msgCtx.msg.data["number"],
	)
}

func msgTooManyArguments(msgCtx MessageContext) error {
	return Exitf(
		exitUsage,
		"error: '%s' accepts at most %d argument(s), but got %s.",
		msgCtx.msg.command.Name(),
		msgCtx.msg.command.MaxArg(),
		msgCtx.msg.data["number"],
	)
}

func getUsage(msgCtx *MessageContext) string {
	var usageMsg string
	helpFlag := msgCtx.app.Config().HelpFlag

	if helpFlag != nil {
		h := helpFlag.Name()
		if h != "" {
			usageMsg = fmt.Sprintf("use --%s for usage information.", h)
		}
	}
	return usageMsg
}

func (a *App) exit(cliMsg *CLIMessage) error {
	getWriter := func(code int) io.Writer {
		if code == exitOK {
			return a.Stdout()
		}
		return a.Stderr()
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
		msgCtx.msg.message = message

		if fn, ok := a.config.CustomMessages[messageType]; fn != nil && ok {
			if err := fn(msgCtx); err != nil {
				message, code = getMessageInfo(err, code)
			}
		}
	}

	out = getWriter(code)
	return newCLIMessage(code, message, messageType, command, data, out)
}

func (a *App) exitWithMsg(messageType messageType, command CommandInfo, data map[string]string) error {
	if data == nil {
		data = map[string]string{}
	}
	return a.exit(newCLIMessage(0, "", messageType, command, data, nil))
}

func (a *App) exitWithErr(err error, defaultCode int) error {
	if e, ok := err.(*CLIMessage); ok {
		defaultCode = e.code
	}
	return a.exit(newCLIMessage(defaultCode, err.Error(), msgNone, nil, nil, nil))
}
