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
	app.Stdout = w

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
	app.Stdout = w

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
	app := &App{
		Name:    "mycli",
		Version: "1.0.0",
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}

	app.Commands = []Command{
		{
			Name:   "message",
			Alias:  "msg",
			Short:  "Send a message",
			Long:   "Send a message to someone",
			MinArg: 1,
			MaxArg: 1,
			Options: []Option{
				{Name: "to", Alias: "t", Value: "", Desc: "Name to send a message to"},
			},
			Run: func(args []string, options map[string]Value) {
				name := options["to"].String()
				text := args[0]

				if name != "" {
					fmt.Fprintf(app.stdout(), "Hey %s! %s\n", name, text)
				} else {
					fmt.Fprintf(app.stdout(), "Hey Guest! %s\n", text)
				}
			},
		},
		{
			Name:  "math",
			Short: "Perform simple math operations",
			Long:  "Perform addition and multiplication operations on numbers",
			Subcommand: []Command{
				{
					Name:   "add",
					Short:  "Adds two numbers",
					MinArg: 2,
					MaxArg: 2,
					Run: func(args []string, options map[string]Value) {
						a := args[0]
						b := args[1]
						fmt.Fprintf(app.stdout(), "%s + %s = %d\n", a, b, atoi(a)+atoi(b))
					},
				},
				{
					Name:   "mul",
					Short:  "Multiplies two numbers",
					MinArg: 2,
					MaxArg: 2,
					Run: func(args []string, options map[string]Value) {
						a := args[0]
						b := args[1]
						fmt.Fprintf(app.stdout(), "%s * %s = %d\n", a, b, atoi(a)*atoi(b))
					},
				},
			},
		},
	}

	return app
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
