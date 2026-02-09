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
	app := NewApp("mycli").SetVersion("0.1.0")

	app.SetStdout(io.Discard)
	app.SetStderr(io.Discard)

	app.AddCommand("message").
		SetAlias("msg").
		SetShort("Send a message").
		SetLong("Send a message to someone").
		SetMinArg(1).
		SetMaxArg(1).
		AddOption("to").SetAlias("t").SetType(String).SetValue("").SetDescription("Name to send a message to").Ok().
		Action(func(args []string, options map[string]Value) {
			name := options["to"].String()
			text := args[0]

			if name != "" {
				fmt.Fprintf(app.Stdout(), "Hey %s! %s\n", name, text)
			} else {
				fmt.Fprintf(app.Stdout(), "Hey Guest! %s\n", text)
			}
		})

	app.AddCommand("math").
		SetShort("Perform simple math operations").
		SetLong("Perform addition and multiplication operations on numbers").
		AddSubcommand("add").
		SetShort("Adds two numbers").
		SetMinArg(2).
		SetMaxArg(2).
		Action(func(args []string, options map[string]Value) {
			a := args[0]
			b := args[1]
			fmt.Fprintf(app.Stdout(), "%s + %s = %d\n", a, b, atoi(a)+atoi(b))
		}).
		Ok().
		AddSubcommand("mul").
		SetShort("Multiplies two numbers").
		SetMinArg(2).
		SetMaxArg(2).
		Action(func(args []string, options map[string]Value) {
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
