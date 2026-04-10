package gocli

// AppConfig represents the configuration settings for the application,
// including default flags and a custom message map.
type AppConfig struct {
	HelpFlag    FlagInfo    // Default help flag
	VersionFlag FlagInfo    // Default version flag
	Messages    MessagesMap // Custom messages map
}

type MessagesMap map[messageType]func(msgCtx MessageContext) error

// DefaultAppConfig creates and returns the default configuration settings.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		HelpFlag: DefaultHelpFlag(),
		Messages: MessagesMap{},
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
