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
		{"global no command", []string{}, "Usage:", 2},
		{"global help flag", []string{"--help"}, "Usage:", 0},
		{"global version flag", []string{"--version"}, "mycli version 0.1.0", 0},
		{"global invalid command", []string{"asd"}, "error: unknown command: 'asd'", 2},
		{"command help menu 1", []string{"message", "-h"}, "Usage", 0},
		{"command help menu 2", []string{"math", "--help"}, "Usage", 0},
		{"command alias", []string{"msg", "Hello"}, "Hey Guest! Hello", 0},
		{"flag variable", []string{"message", "Hi"}, "Hey Guest! Hi", 0},
		{"flag alias", []string{"message", "-t", "John", "Welcome"}, "Hey John! Welcome", 0},
		{"combined flag alias and value", []string{"message", "-tJohn", "Welcome"}, "Hey John! Welcome", 0},
		{"combined flags", []string{"message", "-vtJohn", "Welcome"}, "Hey John! Welcome", 0},
		{"long flag with space", []string{"message", "How are you doing", "--to", "Emily"}, "Hey Emily! How are you doing", 0},
		{"long flag with equal sign", []string{"message", "Hi", "--to=Ben"}, "Hey Ben! Hi", 0},
		{"min argument failure", []string{"message", "--to=Ben"}, "error: 'message' requires at least 1 argument(s), but got 0.", 2},
		{"max argument failure", []string{"message", "--to=Ben", "Hello", "extra"}, "error: 'message' accepts at most 1 argument(s), but got 2.", 2},
		{"missing subcommand", []string{"math"}, "error: a subcommand is required for the command: 'math'", 2},
		{"invalid subcommand", []string{"math", "asd"}, "error: unknown command: 'asd'", 2},
		{"multiple arguments", []string{"math", "add", "2", "2"}, "2 + 2 = 4", 0},
		{"negative value as argument", []string{"math", "add", "-2", "2"}, "error: invalid flag: '-2'", 2},
		{"(--) seperator support", []string{"math", "add", "--", "-2", "2"}, "-2 + 2 = 0", 0},
		{"string slice", []string{"math", "mul", "--numbers=1,2,3", "-n4,5"}, "120", 0},
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

	app.AddGlobalFlag(NewBoolFlag("verbose", false).WithAlias("v").WithDescription("Verbose output"))

	app.WithStdout(io.Discard)
	app.WithStderr(io.Discard)

	defaultName := "Guest"

	messageCmd := NewCommand("message").
		WithAlias("msg").
		WithShort("Send a message").
		WithLong("Send a message to someone").
		WithMinArg(1).
		WithMaxArg(1).
		AddFlag(NewStringFlagVar("to", &defaultName).WithAlias("t").WithDescription("Name to send a message to")).
		WithAction(func(ctx *Context) error {
			name := ctx.String("to")
			text := ctx.Args()[0]

			fmt.Fprintf(app.Stdout(), "Hey %s! %s\n", name, text)
			return nil
		})

	mathCmd := NewCommand("math").
		WithShort("Perform simple math operations").
		WithLong("Perform addition and multiplication operations on numbers")

	mathCmd.
		AddSubcommand(
			NewCommand("add").
				WithShort("Adds two numbers").
				WithMinArg(2).
				WithMaxArg(2).
				WithAction(func(ctx *Context) error {
					a := ctx.Args()[0]
					b := ctx.Args()[1]
					fmt.Fprintf(app.Stdout(), "%s + %s = %d\n", a, b, atoi(a)+atoi(b))
					return nil
				}))

	mathCmd.
		AddSubcommand(
			NewCommand("mul").
				WithShort("Multiplies numbers").
				AddFlag(NewStringSliceFlag("numbers", []string{}).WithAlias("n").WithDescription("Number list to multiply")).
				WithAction(func(ctx *Context) error {
					result := 1
					for _, n := range ctx.StringSlice("numbers") {
						result *= atoi(n)
					}
					fmt.Fprint(app.Stdout(), fmt.Sprint(result)+"\n")
					return nil
				}))

	app.AddCommand(messageCmd)
	app.AddCommand(mathCmd)

	return app
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
