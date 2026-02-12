package gocli

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestApp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		out  string
		code int
	}{
		{"global no command", []string{}, "Usage:", 0},
		{"global help flag", []string{"--help"}, "Usage:", 0},
		{"global version flag", []string{"--version"}, "mycli version 0.1.0", 0},
		{"global invalid command", []string{"asd"}, "unknown command: 'asd'", 2},
		{"command help menu 1", []string{"message", "-h"}, "Usage", 0},
		{"command help menu 2", []string{"math", "--help"}, "Usage", 0},
		{"command alias", []string{"msg", "Hello"}, "Hey Guest! Hello", 0},
		{"flag alias", []string{"message", "-t", "John", "Welcome"}, "Hey John! Welcome", 0},
		{"long flag with space", []string{"message", "How are you doing", "--to", "Emily"}, "Hey Emily! How are you doing", 0},
		{"long flag with equal sign", []string{"message", "Hi", "--to=Ben"}, "Hey Ben! Hi", 0},
		{"min argument failure", []string{"message", "--to=Ben"}, "message requires at least 1 argument(s), got 0", 2},
		{"max argument failure", []string{"message", "--to=Ben", "Hello", "extra"}, "message requires at most 1 argument(s), got 2", 2},
		{"default flag value", []string{"message", "Hi"}, "Hey Guest! Hi", 0},
		{"missing subcommand", []string{"math"}, "math requires a subcommand", 2},
		{"invalid subcommand", []string{"math", "asd"}, "math requires a subcommand", 2},
		{"multiple arguments", []string{"math", "add", "2", "2"}, "2 + 2 = 4", 0},
		{"negative argument value", []string{"math", "add", "-2", "2"}, "invalid flag: '-2'", 2},
		{"(--) seperator support", []string{"math", "add", "--", "-2", "2"}, "-2 + 2 = 0", 0},
	}

	app := exampleApp()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			var out bytes.Buffer
			w := io.MultiWriter(io.Discard, &out)
			app.stdout = w
			app.stderr = w

			args := append([]string{"mycli"}, tc.args...)
			code := app.RunWithArgs(args)

			sout := out.String()

			if code != tc.code {
				t.Fatalf("Expected exit code: %v, but got: %v", tc.code, code)
			}
			if !strings.HasPrefix(sout, tc.out) {
				t.Fatalf("Expected output: %v, but got: %v", tc.out, sout)
			}
		})
	}
}

func exampleApp() *App {
	app := NewApp("mycli").WithVersion("0.1.0")

	app.WithStdout(io.Discard)
	app.WithStderr(io.Discard)

	app.AddCommand("message").
		WithAlias("msg").
		WithShort("Send a message").
		WithLong("Send a message to someone").
		WithMinArg(1).
		WithMaxArg(1).
		AddFlag("to").WithAlias("t").WithDefault("").WithDescription("Name to send a message to").Ok().
		Action(func(args []string, flags Flags) {
			name := flags.String("to")
			text := args[0]

			if name != "" {
				fmt.Fprintf(app.Stdout(), "Hey %s! %s\n", name, text)
			} else {
				fmt.Fprintf(app.Stdout(), "Hey Guest! %s\n", text)
			}
		})

	app.AddCommand("math").
		WithShort("Perform simple math operations").
		WithLong("Perform addition and multiplication operations on numbers").
		AddSubcommand("add").
		WithShort("Adds two numbers").
		WithMinArg(2).
		WithMaxArg(2).
		Action(func(args []string, flags Flags) {
			a := args[0]
			b := args[1]
			fmt.Fprintf(app.Stdout(), "%s + %s = %d\n", a, b, atoi(a)+atoi(b))
		}).
		Ok().
		AddSubcommand("mul").
		WithShort("Multiplies two numbers").
		WithMinArg(2).
		WithMaxArg(2).
		Action(func(args []string, flags Flags) {
			a := args[0]
			b := args[1]
			fmt.Fprintf(app.Stdout(), "%s * %s = %d\n", a, b, atoi(a)*atoi(b))
		}).
		Ok()

	return app
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
