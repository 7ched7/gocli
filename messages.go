package gocli

import "fmt"

type messageType int

const (
	MsgHelp messageType = iota
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
// It includes an exit code, message, message type, and optional command pointer,
// as well as extra metadata.
type CLIMessage struct {
	code        int
	message     string
	messageType messageType
	command     *Command
	data        map[string]string
}

// Code returns the exit code associated with the event.
func (m *CLIMessage) Code() int { return m.code }

// Message returns the message of the event.
func (m *CLIMessage) Message() string { return m.message }

// MessageType returns the categorized type of the event.
func (m *CLIMessage) MessageType() messageType { return m.messageType }

// Command returns the command where the event occured.
func (m *CLIMessage) Command() *Command { return m.command }

// Data returns a map of metadata related to the event.
func (m *CLIMessage) Data() map[string]string { return m.data }

// MessageContext holds the application instance
// and message object providing details about the event.
type MessageContext struct {
	app *App
	msg *CLIMessage
}

// App returns the application instance.
func (m *MessageContext) App() *App { return m.app }

// Msg returns the message object providing details about the event.
func (m *MessageContext) Msg() *CLIMessage { return m.msg }

func (a *App) getMessageAndExitCode(messageType messageType, cmd *Command, data map[string]string) (string, int) {
	name := func() string {
		if cmd == nil || cmd == a.root {
			return a.root.name
		}
		return cmd.name
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
		return fmt.Sprintf("error: unknown command: '%s'\nuse --help for usage information.\n", data["command"]), exitUsage

	case MsgSubcommandRequired:
		return fmt.Sprintf("error: a subcommand is required for the command: '%s'\nuse --help for usage information.\n", data["command"]), exitUsage

	case MsgInvalidFlag:
		return fmt.Sprintf("error: invalid flag: '%s'\nuse --help for usage information.\n", data["flag"]), exitUsage

	case MsgFlagValueMissing:
		return fmt.Sprintf("error: a value is required for the flag: '%s'\n", data["flag"]), exitUsage

	case MsgFlagRequired:
		return fmt.Sprintf("error: flag is required: '%s'\n", data["flag"]), exitUsage

	case MsgUnexpectedArgument:
		return fmt.Sprintf("error: unexpected argument: '%s'\n'%s' does not accept arguments.\n", data["argument"], name()), exitUsage

	case MsgTooFewArguments:
		return fmt.Sprintf("error: '%s' requires at least %d argument(s), but got %s.\n", name(), cmd.minArg, data["number"]), exitUsage

	case MsgTooManyArguments:
		return fmt.Sprintf("error: '%s' accepts at most %d argument(s), but got %s.\n", name(), cmd.maxArg, data["number"]), exitUsage

	case MsgUnsupportedFlagType:
		return fmt.Sprintf("internal error: unsupported flag type: '%s'\n", data["value"]), exitError

	default:
		return "internal error: an unexpected error occurred.\n", exitError
	}
}

func (a *App) stop(messageType messageType, cmd *Command, data map[string]string) int {
	message, code := a.getMessageAndExitCode(messageType, cmd, data)

	cliMsg := &CLIMessage{
		code:        code,
		messageType: messageType,
		command:     cmd,
		message:     message,
		data:        data,
	}

	msgCtx := MessageContext{
		app: a,
		msg: cliMsg,
	}

	return a.displayMessage(msgCtx)
}

func (a *App) displayMessage(msgCtx MessageContext) int {
	var msg string
	isCustom := false
	out := a.Stderr()

	if a.customMessagesMap != nil {
		if fn, ok := a.customMessagesMap[msgCtx.msg.messageType]; ok {
			msg = fn(msgCtx) // override the default message
			isCustom = true
		}
	}

	if !isCustom {
		msg = msgCtx.msg.message
	}

	if msgCtx.msg.code == exitOK {
		out = a.Stdout()
	}

	fmt.Fprint(out, msg)
	return msgCtx.msg.code
}

// HandleMessage registers a custom message handler for a specific message type.
// The handler function runs whenever an event of the given type occurs.
func (a *App) HandleMessage(messageType messageType, fn func(msgCtx MessageContext) string) *App {
	a.customMessagesMap[messageType] = fn
	return a
}
