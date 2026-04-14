![License: MIT](https://img.shields.io/badge/license-MIT-blue)
![Test Status](https://github.com/7ched7/gocli/actions/workflows/test.yml/badge.svg)

# gocli 🦦
A lightweight, dependency-free CLI framework for Go. 

## Overview
**gocli** provides a fluent interface that allows you to define commands, subcommands, and flags in a single, readable format, unlike standard flag parsing. It handles help menu generation, alias mapping, argument validation, error handling, and many other things right out of the box, allowing you to focus on your business logic.

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
	// Create a new App instance
	app := gocli.NewApp("mycli").WithVersion("0.1.0")

	nameFlag := gocli.NewStringFlag("name", "Guest").WithAlias("n")

	app.
		AddGlobalFlag(nameFlag).
		WithMinArg(1).
		WithMaxArg(1).
		WithAction(func(ctx *gocli.Context) error {
			name := ctx.String("name")
			message := ctx.Args()[0]

			return gocli.Exitf(0, "Hey %s! %s", name, message)
		})

	os.Exit(app.Run())
}
```

**Execution Example**
```console
$ mycli -nJohn "How are you?"
Hey John! How are you?
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
		fmt.Println("Server is up and running!")
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
Server is up and running!
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

ipFlag := gocli.NewStringFlagVar("ip", &ip).WithAlias("i")

startCmd.AddFlag(ipFlag).
	WithAction(func(ctx *gocli.Context) error {
		fmt.Printf("IP address: %s\n", ip) // Direct access
		return nil
	})
```

#### Dynamic Access via Context
Another approach is to retrieve the flag value from the **Context** during execution. This is more flexible than binding approach and can be used when multiple flags are involved.
```go
portFlag := gocli.NewIntFlag("port", 0).WithAlias("p")

startCmd.AddFlag(portFlag).
	WithAction(func(ctx *gocli.Context) error {
		port := ctx.Int("port") // Get flag value from context
		fmt.Printf("IP address: %s\n", ip)
		fmt.Printf("Port: %d\n", port)
		return nil
	})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 8000
IP address: 127.0.0.1
Port: 8000
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
ip := ctx.String("ip")

if ip == "127.0.0.1" && value == 8080 {
	return fmt.Errorf("port already in use: %s:%d", ip, value)
}
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
	return fmt.Errorf("invalid IP address: %s", value)
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
ipFlag := gocli.NewCustomFlagVar("ip", &ip).WithAlias("i")
```

**Execution Example**
```console
$ mycli server start -i 256.168.1.1
invalid IP address: 256.168.1.1
```

### Configurations
This framework offers a flexible configuration system that allows you to customize core behaviors.

#### Customizing Default Flags
By default, framework comes with standard flags like **--help** and **--version**. You are free to customize them to better suit your own style and needs.
```go
conf := gocli.DefaultAppConfig()

// Override the version flag alias (e.g., using -V instead of -v)
conf.VersionFlag = gocli.DefaultVersionFlag().WithAlias("V")

app.WithConfig(conf)
```

**Execution Example**
```console
$ mycli -V
mycli version 0.1.0
```

#### Customizing System Messages
One of the core features of the framework is the ability to override default system messages. Using `CustomMessages`, you can provide a more user-friendly output.	 

Simply define a **MessagesMap** and assign a custom function to the specific message type you want to override.
```go
conf.CustomMessages = gocli.MessagesMap{
	gocli.MsgUnknownCommand: func(msgCtx gocli.MessageContext) error {
		return fmt.Errorf("invalid command")
	},
}
```

**Execution Example**
```console
$ mycli servr
invalid command
```

For more advanced scenarios, the **MessageContext** gives you access to detailed runtime state. This makes it possible to generate dynamic messages.
```go
unkCommand := msgCtx.Msg().Data()["command"] // Access the entered invalid command
commands := msgCtx.App().Commands() // All registered commands in app instance

var sb strings.Builder

sb.WriteString("invalid command: ")
sb.WriteString(unkCommand)
sb.WriteString("\navailable commands:\n")

for _, cmd := range commands {
	sb.WriteString("- ")
	sb.WriteString(cmd.Name())
	sb.WriteString("\n")
}

return fmt.Errorf(sb.String())
```

**Execution Example**
```console
$ mycli servr
invalid command: servr
available commands:
- server
```

See the [documentation](https://pkg.go.dev/github.com/7ched7/gocli) for more information.
