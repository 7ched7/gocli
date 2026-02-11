package gocli

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"
)

func TestMessageCommand_WithoutName(t *testing.T) {
	app := exampleApp()

	var out bytes.Buffer
	w := io.MultiWriter(io.Discard, &out)
	app.stdout = w

	app.RunWithArgs([]string{"mycli", "message", "hello"})

	sout := out.String()
	if sout != "Hey Guest! hello\n" {
		t.Fatalf("unexpected output: %s", sout)
	}
}

func TestMessageCommand_WithoutArg(t *testing.T) {
	app := exampleApp()
	c := app.RunWithArgs([]string{"mycli", "message", "--to", "someone"})

	if c != 2 {
		t.Fatalf("expected exit code 2 for missing argument, got %d", c)
	}
}

func TestMathMulCommand(t *testing.T) {
	app := exampleApp()

	var out bytes.Buffer
	w := io.MultiWriter(io.Discard, &out)
	app.stdout = w

	app.RunWithArgs([]string{"mycli", "math", "mul", "2", "2"})

	sout := out.String()
	if sout != "2 * 2 = 4\n" {
		t.Fatalf("unexpected output: %s", sout)
	}
}

func TestMathAddCommand_WithExtraArg(t *testing.T) {
	app := exampleApp()
	c := app.RunWithArgs([]string{"mycli", "math", "add", "3", "4", "6"})

	if c != 2 {
		t.Fatalf("expected exit code 2 for extra argument, got %d", c)
	}
}

func TestMathCommand_WithoutSubcmd(t *testing.T) {
	app := exampleApp()
	c := app.RunWithArgs([]string{"mycli", "math"})

	if c != 2 {
		t.Fatalf("expected exit code 2 for missing subcommand, got %d", c)
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
