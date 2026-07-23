package gocli

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestApp_DeveloperErrors(t *testing.T) {
	tests := []struct {
		name     string
		buildApp func() *App
		out      string
	}{
		{
			name: "empty command name",
			buildApp: func() *App {
				app := NewApp("mycli")
				app.AddCommand(NewCommand(""))
				return app
			},
			out: "command \"mycli\": command name cannot be empty",
		},
		{
			name: "duplicate flag name",
			buildApp: func() *App {
				app := NewApp("mycli")
				cmd := NewCommand("test").
					AddFlag(NewStringFlag("to", "")).
					AddFlag(NewStringFlag("to", ""))
				app.AddCommand(cmd)
				return app
			},
			out: "command \"test\": duplicate flag name \"to\"",
		},
		{
			name: "flag shorthand character long",
			buildApp: func() *App {
				app := NewApp("mycli")
				cmd := NewCommand("test").
					AddFlag(NewStringFlag("", "").WithShorthand("to")).
					AddFlag(NewStringFlag("", "").WithShorthand("to"))
				app.AddCommand(cmd)
				return app
			},
			out: "command \"test\": flag shorthand \"to\" cannot be longer than 1 character",
		},
		{
			name: "invalid argument range",
			buildApp: func() *App {
				app := NewApp("mycli")
				cmd := NewCommand("test").AddArgument(NewArgument("target").WithRange(3, 1))
				app.AddCommand(cmd)
				return app
			},
			out: "command \"test\": minimum value of argument \"target\" cannot be greater than maximum value",
		},
		{
			name: "argument after variadic argument",
			buildApp: func() *App {
				app := NewApp("mycli")
				cmd := NewCommand("test").
					AddArgument(NewArgument("target").WithRange(1, -1)).
					AddArgument(NewArgument("source"))
				app.AddCommand(cmd)
				return app
			},
			out: "command \"test\": argument \"source\" cannot follow a variadic argument",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Sprintf("%v", r)
					if !strings.Contains(err, tc.out) {
						t.Fatalf("expected output: %q, got: %q", tc.out, err)
					}
				}
			}()

			app := tc.buildApp()
			app.RunWithArgs([]string{"mycli", "test"})
		})
	}
}

func TestApp_UserErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		out  string
		code int
	}{
		{"global no command", []string{}, "Usage:", 0},
		{"global help flag", []string{"--help"}, "Usage:", 0},
		{"global version flag", []string{"--version"}, "mycli version 0.1.0", 0},
		{"global invalid command", []string{"asd"}, "error: unknown command: 'asd'", 2},
		{"command help menu 1", []string{"message", "-h"}, "Usage", 0},
		{"command help menu 2", []string{"math", "--help"}, "Usage", 0},
		{"command alias", []string{"msg", "hello"}, "hey guest! hello", 0},
		{"flag variable", []string{"message", "hi"}, "hey guest! hi", 0},
		{"flag shorthand", []string{"message", "-t", "john", "welcome"}, "hey john! welcome", 0},
		{"combined flag shorthand and value", []string{"message", "-tjohn", "welcome"}, "hey john! welcome", 0},
		{"combined flags", []string{"message", "-Vtjohn", "welcome"}, "hey john! welcome", 0},
		{"long flag with space", []string{"message", "how are you doing", "--to", "emily"}, "hey emily! how are you doing", 0},
		{"long flag with equal sign", []string{"message", "hi", "--to=ben"}, "hey ben! hi", 0},
		{"flag parse error", []string{"--verbose=tru"}, "error: invalid value 'tru' for flag '--verbose': expected boolean", 2},
		{"unexpected argument", []string{"math", "mul", "0"}, "error: unexpected argument: '0'", 2},
		{"min argument failure", []string{"message", "--to=Ben"}, "error: argument 'message' expects at least 1 value(s), got 0", 2},
		{"max argument failure", []string{"message", "--to=Ben", "Hello", "extra"}, "error: argument 'message' expects at most 1 value(s), got 2", 2},
		{"missing subcommand", []string{"math"}, "error: command 'math' requires a subcommand", 2},
		{"invalid subcommand", []string{"math", "asd"}, "error: unknown command: 'asd'", 2},
		{"multiple arguments", []string{"math", "add", "2", "2"}, "2 + 2 = 4", 0},
		{"negative value as argument", []string{"math", "add", "-2", "2"}, "error: invalid flag: '-2'", 2},
		{"(--) seperator support", []string{"math", "add", "--", "-2", "2"}, "-2 + 2 = 0", 0},
		{"string slice", []string{"math", "mul", "--numbers=1,2,3", "-n4,5"}, "120", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := buildApp()

			var out bytes.Buffer
			w := io.MultiWriter(io.Discard, &out)

			args := append([]string{"mycli"}, tc.args...)
			err := app.RunWithArgs(args)

			code := 0
			if err != nil {
				fmt.Fprint(w, err)
				code = err.(*CLIMessage).code
			}

			sout := out.String()

			if code != tc.code {
				t.Fatalf("expected exit code: %v, got: %v", tc.code, code)
			}
			if !strings.Contains(sout, tc.out) {
				t.Fatalf("expected output: %v, got: %v", tc.out, sout)
			}
		})
	}
}

func buildApp() *App {
	app := NewApp("mycli").WithVersion("0.1.0")

	app.AddGlobalFlag(NewBoolFlag("verbose", false).WithShorthand("V"))

	defaultName := "guest"

	messageCmd := NewCommand("message").
		WithAlias("msg").
		AddArgument(NewArgument("message")).
		AddFlag(NewStringFlagVar("to", &defaultName).WithShorthand("t")).
		WithAction(func(ctx *Context) error {
			name := ctx.Flag("to").String()
			text := ctx.Arg("message").Get(0)
			return Exitf(0, "hey %s! %s\n", name, text)
		})

	mathCmd := NewCommand("math").
		AddSubcommand(
			NewCommand("add").
				AddArgument(NewArgument("number").WithRange(2, 2)).
				WithAction(func(ctx *Context) error {
					numbers := ctx.Arg("number").All()
					a := numbers[0]
					b := numbers[1]
					return Exitf(0, "%s + %s = %d\n", a, b, atoi(a)+atoi(b))
				})).
		AddSubcommand(
			NewCommand("mul").
				AddFlag(NewStringSliceFlag("numbers", []string{}).WithShorthand("n")).
				WithAction(func(ctx *Context) error {
					result := 1
					for _, n := range ctx.Flag("numbers").StringSlice() {
						result *= atoi(n)
					}
					return Exitf(0, "%d", result)
				}))

	app.AddCommand(messageCmd, mathCmd)
	return app
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
