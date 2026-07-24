![License: MIT](https://img.shields.io/badge/license-MIT-blue)
![Test Status](https://github.com/7ched7/gocli/actions/workflows/test.yml/badge.svg)

# gocli 🦦
A lightweight, dependency-free CLI framework for Go. 

## Overview
**gocli** provides a fluent interface that lets you define commands and flags in a single, readable format. It handles help menu generation, alias mapping, argument validation, error handling, and many other things right out of the box, allowing you to focus on your business logic.

## Installation
```bash
go get github.com/7ched7/gocli@latest
```

## Quick Start
```go 
package main

import (
	"os"

	"github.com/7ched7/gocli"
)

func main() {
	// Create a new app instance
	app := gocli.NewApp("mycli").WithVersion("0.1.0")

	// Define flags and arguments
	nameFlag := gocli.NewStringFlag("name", "guest").WithShorthand("n")
	messageArg := gocli.NewArgument("message")

	// Configure and run the application
	app.
		AddGlobalFlag(nameFlag).
		AddArgument(messageArg).
		WithAction(func(ctx *gocli.Context) error {
			name := ctx.Flag("name").String()
			message := ctx.Arg("message").Get(0)

			return gocli.Exitf(0, "hey %s! %s", name, message)
		})

	os.Exit(app.Run())
}
```

**Execution Example**
```console
$ mycli -njohn "how are you?"
hey john! how are you?
```

## Features
### Commands & Subcommands
This framework lets you define standalone commands and organize them hierarchically using subcommands.

To create a new command, you can simply use the `NewCommand` method provided by the API. 
```go
serverCmd := gocli.NewCommand("server")
```

Once defined, register the command by passing it to the `AddCommand` method of the app instance.
```go
app.AddCommand(serverCmd)
```

If your project requires nested commands, you can register the command directly within a parent command using the `AddSubcommand` method.
```go
startCmd := gocli.NewCommand("start").
	WithAction(func(ctx *gocli.Context) error {
		fmt.Println("server is up and running!")
		return nil
	})

serverCmd.AddSubcommand(startCmd)
```

If you don't want to mess around with variables, simply call the `NewCommand` method directly inside the `AddCommand` method. This is the another way also.
```go
app.AddCommand(gocli.NewCommand("server"))
```

**Execution Example**
```console
$ mycli server start 
server is up and running!
```

### Flags
Flags are created using type-specific helper methods and then registered on a command using the `AddFlag` method.
```go
ipFlag := gocli.NewStringFlag("ip", "")
startCmd.AddFlag(ipFlag)
```

You can also add flags globally. Simply pass the flag to the `AddGlobalFlag` method of the app instance. No matter which command you run, the global flags will always be available. 
```go
verboseFlag := gocli.NewBoolFlag("verbose", false)
app.AddGlobalFlag(verboseFlag)
```

In this framework, flags are accessed in two main ways:
- [Variable Binding](#variable-binding) - bind the variable to the flag
- [Dynamic Access via Context](#dynamic-access-via-context) - access the flag at runtime

#### Variable Binding
With variable binding, the flag value is automatically stored in a predefined variable. This is useful when you want simple and direct access to the value.
```go
var ip string = "127.0.0.1"

ipFlag := gocli.NewStringFlagVar("ip", &ip).WithShorthand("i")

startCmd.AddFlag(ipFlag).
	WithAction(func(ctx *gocli.Context) error {
		fmt.Printf("ip address: %s\n", ip) // Direct access
		return nil
	})
```

#### Dynamic Access via Context
Another approach is to retrieve the flag value from the **Context** during execution. This is more flexible than binding approach and can be used when multiple flags are involved.
```go
portFlag := gocli.NewIntFlag("port", 0).WithShorthand("p")

startCmd.AddFlag(portFlag).
	WithAction(func(ctx *gocli.Context) error {
		port := ctx.Flag("port").Int() // Get flag value from context
		fmt.Printf("ip address: %s\n", ip)
		fmt.Printf("port: %d\n", port)
		return nil
	})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 8000
ip address: 127.0.0.1
port: 8000
```

### Custom Validator
Flags may need to be validated to ensure they meet specific conditions. In such cases, you can use the `WithValidator` method on the flag to validate the flag value just before the action is triggered.
```go
portFlag.WithValidator(func(ctx *gocli.Context, value int) error {
	if value < 1 || value > 65535 {
		return fmt.Errorf("invalid port number: %d", value)
	}
	return nil
})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 80000
invalid port number: 80000
```

The **Context** allows you to write more complex controls by providing access to other flag values.
```go
portFlag.WithValidator(func(ctx *gocli.Context, value int) error {
	ip := ctx.Flag("ip").String() // Access the other flag value

	if ip == "127.0.0.1" && value == 8080 {
		return fmt.Errorf("port already in use: %s:%d", ip, value)
	}
	return nil
})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 8080
port already in use: 127.0.0.1:8080
```

### Custom Types
This framework doesn’t support every type out of the box. Instead, it lets you define and integrate custom types by implementing the **FlagValue** interface, which closely mirrors Go’s standard [flag.Value](https://pkg.go.dev/flag#Value).

The first step is to wrap your desired data type within a struct. This struct acts as a container that the framework will interact with during the flag lifecycle.
```go
type IP struct {
	value net.IP
}
```

The `Set(string)` method is the core of your custom type. The framework calls this method whenever it encounters your flag in the CLI. It is responsible for converting the raw string input into your internal data type.
```go
func (i *IP) Set(value string) error {
	if ip := net.ParseIP(value); ip != nil {
		i.value = ip
		return nil
	}
	return fmt.Errorf("invalid ip address: %s", value)
}
```

To retrieve the processed data back, you need to implement the `Get()` method. This returns the underlying value as **any** type.
```go
func (i *IP) Get() any { 
	return i.value 
}
```

The `String()` method is implemented to satisfy the [Stringer](https://go.dev/tour/methods/17) interface. It represents either the current value or a hint about the expected format.
```go
func (i *IP) String() string { 
	return i.value.String() 
}
```

Once your struct satisfies the **FlagValue** interface, you can integrate it into the flag using the `NewCustomFlagVar` method provided by the API. Simply create a typed variable and bind it.
```go
var ip IP
ipFlag := gocli.NewCustomFlagVar("ip", &ip).WithShorthand("i")
```

**Execution Example**
```console
$ mycli server start -i 256.168.1.1
invalid ip address: 256.168.1.1
```

### Configurations
This framework offers a flexible configuration system that allows you to customize core behaviors.

#### Customizing Default Flags
By default, framework comes with default flags like **--help** and **--version**. You are free to customize them to better suit your own style and needs.
```go
cfg := gocli.DefaultAppConfig()

// Override the version flag shorthand (e.g., using -V instead of -v)
cfg.VersionFlag = gocli.DefaultVersionFlag().WithShorthand("V")

app.WithConfig(cfg)
```

**Execution Example**
```console
$ mycli -V
mycli version 0.1.0
```

#### Customizing Default Messages
One of the core features of the framework is the ability to customize or override default messages to better suit your application's tone, custom logic, or localization needs.

Simply define a **MessagesMap** and assign a custom function to the specific message type you want to override.
```go
cfg.CustomMessages = gocli.MessagesMap{
	gocli.MsgUnknownCommand: func(msgCtx gocli.MessageContext) error {
		return errors.New("unknown command")
	},
}
```

**Execution Example**
```console
$ mycli servr
unknown command
```

For more advanced scenarios, the **MessageContext** gives you access to detailed runtime state. This makes it possible to generate dynamic messages.
```go
cfg.CustomMessages = gocli.MessagesMap{
	gocli.MsgUnknownCommand: func(msgCtx gocli.MessageContext) error {
		unkCommand := msgCtx.Msg().Data()["command"] // Access the entered unknown command
		commands := msgCtx.App().Commands() // All registered commands in app

		var sb strings.Builder

		sb.WriteString("unknown command: ")
		sb.WriteString(unkCommand)
		sb.WriteString("\navailable commands:")

		for _, cmd := range commands {
			sb.WriteString("\n")
			sb.WriteString("- ")
			sb.WriteString(cmd.Name())
		}

		return errors.New(sb.String())
	},
}
```

**Execution Example**
```console
$ mycli servr
unknown command: servr
available commands:
- server
```

See the [documentation](https://pkg.go.dev/github.com/7ched7/gocli) for more information.
