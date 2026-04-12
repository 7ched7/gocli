package gocli

import (
	"io"
	"os"
)

// AppConfig represents the configuration settings for the application.
// It includes default flags, custom messages map, and I/O writers.
type AppConfig struct {
	HelpFlag       FlagInfo    // Default help flag
	VersionFlag    FlagInfo    // Default version flag
	CustomMessages MessagesMap // Custom messages map
	Stdout         io.Writer   // Output writer used for standard output
	Stderr         io.Writer   // Output writer used for standard error
}

// MessagesMap maps each message type to its handler function.
type MessagesMap map[messageType]func(msgCtx MessageContext) error

// DefaultAppConfig creates and returns the default configuration settings.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		HelpFlag:       DefaultHelpFlag(),
		CustomMessages: MessagesMap{},
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
	}
}

// DefaultHelpFlag creates and returns the default help flag.
func DefaultHelpFlag() *Flag[bool] {
	helpFlag := NewBoolFlag("help", false).WithAlias("h").WithDescription("Show help")
	helpFlag.setRole(flagHelp)
	return helpFlag
}

// DefaultVersionFlag creates and returns the default version flag.
func DefaultVersionFlag() *Flag[bool] {
	versionFlag := NewBoolFlag("version", false).WithAlias("v").WithDescription("Show version")
	versionFlag.setRole(flagVersion)
	return versionFlag
}
