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

// Exit creates a new CLI message with the provided message and code.
func Exit(message string, code int) *CLIMessage {
	return &CLIMessage{message: message, code: code}
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
	return Exit(msgCtx.app.Help(), exitOK)
}

func msgCommandHelp(msgCtx MessageContext) error {
	return Exit(msgCtx.app.CommandHelp(msgCtx.msg.command), exitOK)
}

func msgVersion(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"%s version %s\n",
		msgCtx.app.Name(),
		msgCtx.app.Version(),
	)
	return Exit(m, exitOK)
}

func msgNoCommand(msgCtx MessageContext) error {
	return Exit(msgCtx.App().Help(), exitUsage)
}

func msgUnknownCommand(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: unknown command: '%s'\n%s",
		msgCtx.msg.data["command"],
		getUsage(&msgCtx),
	)
	return Exit(m, exitUsage)
}

func msgSubcommandRequired(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: a subcommand is required for the command: '%s'\n%s",
		msgCtx.msg.data["command"],
		getUsage(&msgCtx),
	)
	return Exit(m, exitUsage)
}

func msgInvalidFlag(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: invalid flag: '%s'\n%s",
		msgCtx.msg.data["flag"],
		getUsage(&msgCtx),
	)
	return Exit(m, exitUsage)
}

func msgFlagValueMissing(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: a value is required for the flag: '%s'\n",
		msgCtx.msg.data["flag"],
	)
	return Exit(m, exitUsage)
}

func msgFlagRequired(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: flag is required: '%s'\n",
		msgCtx.msg.data["flag"],
	)
	return Exit(m, exitUsage)
}

func msgIntParseError(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: invalid value '%v': must be an integer.\n",
		msgCtx.msg.data["value"],
	)
	return Exit(m, exitUsage)
}

func msgFloat64ParseError(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: invalid value '%v': must be a float.\n",
		msgCtx.msg.data["value"],
	)
	return Exit(m, exitUsage)
}

func msgBoolParseError(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: invalid value '%v': must be a bool.\n",
		msgCtx.msg.data["value"],
	)
	return Exit(m, exitUsage)
}

func msgUnexpectedArgument(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: unexpected argument: '%s'\n'%s' does not accept arguments.\n",
		msgCtx.msg.data["argument"],
		msgCtx.msg.command.Name(),
	)
	return Exit(m, exitUsage)
}

func msgTooFewArguments(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: '%s' requires at least %d argument(s), but got %s.\n",
		msgCtx.msg.command.Name(),
		msgCtx.msg.command.MinArg(),
		msgCtx.msg.data["number"],
	)
	return Exit(m, exitUsage)
}

func msgTooManyArguments(msgCtx MessageContext) error {
	m := fmt.Sprintf(
		"error: '%s' accepts at most %d argument(s), but got %s.\n",
		msgCtx.msg.command.Name(),
		msgCtx.msg.command.MaxArg(),
		msgCtx.msg.data["number"],
	)
	return Exit(m, exitUsage)
}

func getUsage(msgCtx *MessageContext) string {
	var usageMsg string
	helpFlag := msgCtx.app.Config().HelpFlag

	if helpFlag != nil {
		h := helpFlag.Name()
		if h != "" {
			usageMsg = fmt.Sprintf("use --%s for usage information.\n", h)
		}
	}
	return usageMsg
}

func (a *App) exit(cliMsg *CLIMessage) int {
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
	out := getWriter(code)

	if messageType == msgNone {
		fmt.Fprint(out, message)
		return code
	}

	msgCtx := MessageContext{
		app: a,
		msg: &CLIMessage{
			code:        code,
			messageType: messageType,
			command:     cliMsg.command,
			data:        cliMsg.data,
		},
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
	fmt.Fprint(out, message)
	return code
}

func (a *App) exitWithMsg(messageType messageType, command CommandInfo, data map[string]string) int {
	if data == nil {
		data = map[string]string{}
	}

	return a.exit(&CLIMessage{
		messageType: messageType,
		command:     command,
		data:        data,
	})
}

func (a *App) exitWithErr(err error, defaultCode int) int {
	if e, ok := err.(*CLIMessage); ok {
		defaultCode = e.code
	}

	return a.exit(&CLIMessage{
		code:        defaultCode,
		message:     err.Error(),
		messageType: msgNone,
	})
}
