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
	config := AppConfig{}

	config.HelpFlag = config.DefaultHelpFlag()
	config.VersionFlag = config.DefaultVersionFlag()
	config.Messages = MessagesMap{}

	return config
}

// DefaultHelpFlag creates and returns the default help flag.
func (AppConfig) DefaultHelpFlag() *Flag[bool] {
	helpFlag := NewBoolFlag("help", false).WithAlias("h").WithDescription("Show help")
	helpFlag.setRole(flagHelp)
	return helpFlag
}

// DefaultVersionFlag creates and returns the default version flag.
func (AppConfig) DefaultVersionFlag() *Flag[bool] {
	versionFlag := NewBoolFlag("version", false).WithAlias("v").WithDescription("Show version")
	versionFlag.setRole(flagVersion)
	return versionFlag
}
