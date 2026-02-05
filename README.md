# Very Simple CLI Parser for Go 🦦
Lightweight, zero dependencies, and useful for simple command parsing. 

## Get Package
```bash
go get github.com/7ched7/gocli@latest
```

## Example Usage
```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/7ched7/gocli"
)

func main() {
	app := &gocli.App{
		Name:    "mycli",
		Version: "1.0.0",
		Commands: []gocli.Command{
			{
				Name:   "message",
				Alias:  "msg",
				Short:  "Send a message",
				Long:   "Send a message to someone",
				MinArg: 1,
				MaxArg: 1,
				Options: []gocli.Option{
					{Name: "to", Alias: "t", Value: "", Desc: "Name to send a message to"},
				},
				Run: func(args []string, options map[string]gocli.Value) {
					name := options["to"].String()
					text := args[0]

					if name != "" {
						fmt.Printf("Hey %s! %s\n", name, text)
					} else {
						fmt.Printf("Hey Guest! %s\n", text)
					}
				},
			},
			{
				Name:  "math",
				Short: "Perform simple math operations",
				Long:  "Perform addition and multiplication operations on numbers",
				Subcommand: []gocli.Command{
					{
						Name:   "add",
						Short:  "Adds two numbers",
						MinArg: 2,
						MaxArg: 2,
						Run: func(args []string, options map[string]gocli.Value) {
							a := args[0]
							b := args[1]
							fmt.Printf("%s + %s = %d\n", a, b, atoi(a)+atoi(b))
						},
					},
					{
						Name:   "mul",
						Short:  "Multiplies two numbers",
						MinArg: 2,
						MaxArg: 2,
						Run: func(args []string, options map[string]gocli.Value) {
							a := args[0]
							b := args[1]
							fmt.Printf("%s * %s = %d\n", a, b, atoi(a)*atoi(b))
						},
					},
				},
			},
		},
	}

	os.Exit(app.Run())
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
```

## Example Commands
```bash
$ mycli --version
# mycli version 1.0.0

$ mycli --help
# Usage:
#   mycli [COMMAND] [OPTIONS] [ARGS...]
#
# Commands:
#   message, msg       Send a message
#   math               Perform simple math operations
#
# Options:
#   --help, -h         Show help
#   --version, -v      Show version

# For more information about a command, use 'mycli <command> --help'.

$ mycli message --help
# Usage:
#   mycli message [OPTIONS] [ARG]
#
# Send a message to someone
#
# Options:
#   --to, -t           Name to send a message to

$ mycli message Welcome
# Hey Guest! Welcome

$ mycli msg --to John "How are you doing?" # The order matters, options come before the arguments 
# Hey John! How are you doing?

$ mycli math --help
# Usage:
#   mycli math [COMMAND]
#
# Perform addition and multiplication operations on numbers
#
# Commands:
#   add                Adds two numbers
#   mul                Multiplies two numbers

$ mycli math add 5 7
# 5 + 7 = 12

$ mycli math mul 3 4
# 3 * 4 = 12
```

This package is not feature-rich, and is tailored to personal needs. If you are looking for a CLI framework with more advanced features, [see](https://github.com/shadawck/awesome-cli-frameworks?tab=readme-ov-file#go) here. 

**[License MIT](LICENSE)**