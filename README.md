![License: MIT](https://img.shields.io/badge/license-MIT-bright)
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
	"fmt"
	"os"

	"github.com/7ched7/gocli"
)

func main() {
	// Create a new App instance
	app := gocli.NewApp("mycli").WithVersion("0.1.0")

	nameFlag := gocli.NewStringFlag("name").WithAlias("n")

	app.
		AddGlobalFlag(nameFlag).
		WithMinArg(1).
		WithMaxArg(1).
		Action(func(ctx *gocli.Context) {
			name := ctx.String("name")
			message := ctx.Args()[0]

			if name != "" {
				fmt.Printf("Hey %s! %s\n", name, message)
			} else {
				fmt.Printf("Hey Guest! %s\n", message)
			}

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

If your project requires nested commands, you can register the command object directly within a parent command object using the `AddSubcommand` method.
```go
startCmd := gocli.NewCommand("start").
	Action(func(ctx *gocli.Context) {
		fmt.Println("Server is up and running!")
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
ipFlag := gocli.NewStringFlag("ip")
startCmd.AddFlag(ipFlag)
```

You can also add flags globally. Simply pass the flag object to the `AddGlobalFlag` method of the app instance. No matter which command you run, the global flags will always be available. 
```go
verboseFlag := gocli.NewBoolFlag("verbose")
app.AddGlobalFlag(verboseFlag)
```

In the framework, flags are accesed in two main ways:
- [Variable Binding](#variable-binding) - bind the variable to the flag
- [Dynamic Access via Context](#dynamic-access-via-context) - access the flag at runtime

#### Variable Binding
With variable binding, the flag value is automatically stored in a predefined variable. This is useful when you want simple and direct access to the value.
```go
var ip string = "127.0.0.1"

ipFlag := gocli.NewStringFlagVar("ip", &ip).WithAlias("i")

startCmd.AddFlag(ipFlag).
	Action(func(ctx *gocli.Context) {
		fmt.Printf("IP address: %s\n", ip) // Direct access
	})
```

#### Dynamic Access via Context
Another approach is to retrieve the flag value from the **Context** during execution. This is more flexible than binding approach and can be used when multiple flags are involved.
```go
portFlag := gocli.NewIntFlag("port").WithAlias("p")

startCmd.AddFlag(portFlag).
	Action(func(ctx *gocli.Context) {
		port := ctx.Int("port") // Get flag value from context
		fmt.Printf("IP address: %s\n", ip)
		fmt.Printf("Port: %d\n", port)
	})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.2 -p 8000
IP address: 127.0.0.2
Port: 8000
```

### Custom Validator
Flags may need to be validated to ensure they meet specific conditions. In such cases, you can use the `WithValidator` method on the flag object to validate the flag value just before the command action is triggered.
```go
portFlag.WithValidator(func(ctx *gocli.Context, value int) error {
	if value < 1 || value > 65535 {
		return fmt.Errorf("invalid port number: %d\n", value)
	}
	return nil
})
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 80000
invalid port number: 80000
```

The **Context** object can be used to write more complex controls by accessing other flag values.
```go
ip := ctx.String("ip")

if ip == "127.0.0.1" && value == 8080 {
	return fmt.Errorf("port already in use: %s:%d\n", ip, value)
}
```

**Execution Example**
```console
$ mycli server start -i 127.0.0.1 -p 8080
port already in use: 127.0.0.1:8080
```

### Custom Type
The framework doesn’t support every type out of the box. Instead, it lets you define and integrate custom types by implementing the **FlagValue** interface, which closely mirrors Go’s standard [flag.Value](https://pkg.go.dev/flag#Value).

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
	return fmt.Errorf("invalid IP address: %s\n", value)
}
```
To retrieve the processed data back from the framework, you need to implement the `Get()` method. This returns the underlying value as **any** type.
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
ipFlag := gocli.NewCustomFlagVar("ip", &ip)
```

**Execution Example**
```console
$ mycli server start -i 256.168.1.1
invalid IP address: 256.168.1.1
```

### Custom Messages
One of the core features of the framework is its flexible message handling. You can override the default behavior for various CLI events to provide more user-friendly messages. 

Simply specify one of the predefined events provided by the framework using the `HandleMessage` method and return the message you want to display.
```go
app.HandleMessage(gocli.MsgUnknownCommand, func(msgCtx gocli.MessageContext) string {
	return "invalid command\n"
})
```

**Execution Example**
```console
$ mycli servr
invalid command
```

The **MessageContext** object can be used to generate more advanced messages. You can access information about the event and other objects.

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

return sb.String()
```

**Execution Example**
```console
$ mycli servr
invalid command: servr
available commands:
- server
```

See the [documentation](https://pkg.go.dev/github.com/7ched7/gocli) for more information.
