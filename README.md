![Test Status](https://github.com/7ched7/gocli/actions/workflows/test.yml/badge.svg)

## gocli 🦦
A lightweight, dependency-free CLI framework for Go. 

### Overview
This framework provides a fluent interface that allows you to define commands, subcommands, and flags in a single, readable format, unlike standard flag parsing. It handles dynamic help menu generation, alias mapping, argument validation, error handling, and many other things right out of the box, allowing you to focus on your business logic.

### Installation

```bash
go get github.com/7ched7/gocli@latest
```

### Example Usage

```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/7ched7/gocli"
)

func main() {
	// Create a new App instance
	app := gocli.NewApp("mycli").WithVersion("0.1.0")

	app.AddCommand("message").
		WithAlias("msg").
		WithShort("Send a message").
		WithLong("Send a message to someone").
		WithMinArg(1).
		WithMaxArg(1).
		AddFlag("to").WithAlias("t").WithDefault("").WithDescription("Name to send a message to").Ok().
		Action(func(args []string, flags gocli.Flags) {
			name := flags.String("to")
			text := args[0]

			if name != "" {
				fmt.Printf("Hey %s! %s\n", name, text)
			} else {
				fmt.Printf("Hey Guest! %s\n", text)
			}
		})

	app.AddCommand("math").
		WithShort("Perform simple math operations").
		WithLong("Perform addition and multiplication operations on numbers").
		AddSubcommand("add").
		WithShort("Adds two numbers").
		WithMinArg(2).
		WithMaxArg(2).
		Action(func(args []string, flags gocli.Flags) {
			a := args[0]
			b := args[1]
			fmt.Printf("%s + %s = %d\n", a, b, atoi(a)+atoi(b))
		}).
		Ok()

	os.Exit(app.Run())
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
```

### Example Commands

**Help Menu**
```bash
$ mycli --help
# Usage:
#   mycli [COMMAND] [FLAGS] [ARGS...]
#
# Commands:
#   message, msg       Send a message
#   math               Perform simple math operations
#
# Flags:
#   --help, -h         Show help
#   --version, -v      Show version
#
# For more information about a command, use 'mycli <command> --help'.

$ mycli message --help
# Usage:
#   mycli message [FLAGS] [ARG]
#
# Send a message to someone
#
# Flags:
#   --to, -t           Name to send a message to
```

**Flags**
```bash
$ mycli message Welcome --to John
# Hey John! Welcome

$ mycli message -t=David Welcome
# Hey David! Welcome
```

**Subcommands**
```bash
$ mycli math add 5 7
# 5 + 7 = 12

$ mycli math add -- -2 3
# -2 + 3 = 1
```

### Customizing Messages
One of the core features of `gocli` is its flexible message handling. You can override the default behavior for various CLI events to provide a more user-friendly messages. Here is a simple example of how to implement it when a specific type of error occurs:

```go
app.HandleMessage(gocli.ErrUnknownCommand, func(app *gocli.App, err gocli.CLIError) string {
	return fmt.Sprintf("'%s' is not a valid command. See '%s --help'.\n", err.Data["command"], app.Name())
})
```

**[License MIT](LICENSE)**
