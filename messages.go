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

// MessageType returns the internal type of the message.
func (m *CLIMessage) MessageType() messageType { return m.messageType }

// Command returns the command that triggered the message.
func (m *CLIMessage) Command() CommandInfo { return m.command }

// Data returns the metadata associated with the message.
func (m *CLIMessage) Data() map[string]string { return m.data }

// Writer returns the writer where the message is written.
func (m *CLIMessage) Writer() io.Writer { return m.writer }

// Exit creates a new CLI message with the provided code and message.
func Exit(code int, message string) *CLIMessage {
	return &CLIMessage{
		code:    code,
		message: message,
	}
}

// Exitf creates a new CLI message with the code and formatted message.
func Exitf(code int, format string, a ...any) *CLIMessage {
	message := fmt.Sprintf(format, a...)
	return &CLIMessage{
		code:    code,
		message: message,
	}
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

// DefaultMessage returns the default system message.
func (m *MessageContext) DefaultMessage() string {
	if fn, ok := defaultMessages[m.msg.messageType]; ok {
		if err := fn(*m); err != nil {
			return err.Error()
		}
	}
	return m.msg.message
}

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
	return fmt.Errorf(msgCtx.app.Help())
}

func msgCommandHelp(msgCtx MessageContext) error {
	return fmt.Errorf(msgCtx.app.CommandHelp(msgCtx.msg.command))
}

func msgVersion(msgCtx MessageContext) error {
	return fmt.Errorf(
		"%s version %s",
		msgCtx.app.Name(),
		msgCtx.app.Version(),
	)
}

func msgNoCommand(msgCtx MessageContext) error {
	return fmt.Errorf(msgCtx.App().Help())
}

func msgUnknownCommand(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: unknown command: '%s'%s",
		msgCtx.msg.data["command"],
		msgUsage(&msgCtx),
	)
}

func msgSubcommandRequired(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: a subcommand is required for command '%s'%s",
		msgCtx.msg.data["command"],
		msgUsage(&msgCtx),
	)
}

func msgInvalidFlag(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: invalid flag: '%s'%s",
		msgCtx.msg.data["flag"],
		msgUsage(&msgCtx),
	)
}

func msgFlagValueMissing(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: a value is required for flag '%s'",
		msgCtx.msg.data["flag"],
	)
}

func msgFlagRequired(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: flag is required: '%s'",
		msgCtx.msg.data["flag"],
	)
}

func msgIntParseError(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: invalid value '%v': expected integer",
		msgCtx.msg.data["value"],
	)
}

func msgFloat64ParseError(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: invalid value '%v': expected float",
		msgCtx.msg.data["value"],
	)
}

func msgBoolParseError(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: invalid value '%v': expected boolean",
		msgCtx.msg.data["value"],
	)
}

func msgUnexpectedArgument(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: unexpected argument: '%s'",
		msgCtx.msg.data["argument"],
	)
}

func msgTooFewArguments(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: argument '%s' expects at least %s value(s), but got %s",
		msgCtx.msg.data["name"],
		msgCtx.msg.data["min"],
		msgCtx.msg.data["got"],
	)
}

func msgTooManyArguments(msgCtx MessageContext) error {
	return fmt.Errorf(
		"error: argument '%s' expects at most %s value(s), but got %s",
		msgCtx.msg.data["name"],
		msgCtx.msg.data["max"],
		msgCtx.msg.data["got"],
	)
}

func msgUsage(msgCtx *MessageContext) string {
	if msgCtx.app.Config().HelpFlag != nil {
		h := flagDisplayName(msgCtx.app.Config().HelpFlag, true)
		return fmt.Sprintf("\nuse '%s' for usage information.", h)
	}
	return ""
}

func (a *App) exit(m *CLIMessage) error {
	cliMsg := *m

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

	cliMsg.writer = getWriter(cliMsg.code)

	if cliMsg.messageType == msgNone {
		return &cliMsg
	}

	msgCtx := MessageContext{
		app: a,
		msg: &cliMsg,
	}

	fn := defaultMessages[cliMsg.messageType]

	if customFn, ok := a.config.CustomMessages[cliMsg.messageType]; ok && customFn != nil {
		fn = customFn
	}

	if fn != nil {
		if err := fn(msgCtx); err != nil {
			cliMsg.message, cliMsg.code = getMessageInfo(err, cliMsg.code)
		}
	}

	cliMsg.writer = getWriter(cliMsg.code)
	return &cliMsg
}

func (a *App) exitWithMsg(code int, messageType messageType, command CommandInfo, data map[string]string) error {
	return a.exit(&CLIMessage{
		code:        code,
		messageType: messageType,
		command:     command,
		data:        data,
	})
}

func (a *App) exitWithErr(code int, err error, command CommandInfo) error {
	if e, ok := err.(*CLIMessage); ok {
		code = e.code
	}
	return a.exit(&CLIMessage{
		code:    code,
		message: err.Error(),
		command: command,
	})
}
