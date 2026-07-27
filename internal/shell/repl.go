package shell

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/minderrx2-art/monsh/internal/builtin"
	"github.com/minderrx2-art/monsh/internal/parser"
	"github.com/minderrx2-art/monsh/internal/path"
	"github.com/minderrx2-art/monsh/internal/runner"
	"github.com/minderrx2-art/monsh/internal/shell/completer"
)

func newReader(builtins *Builtins) (*readline.Instance, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		HistoryFile:     "/tmp/monsh.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete: &completer.ShellCompleter{
			ReadlineCompleter: *readline.NewPrefixCompleter(
				readline.PcItem("exit"),
				readline.PcItem("pwd"),
				readline.PcItem("cd", readline.PcItemDynamic(func(line string) []string {
					return path.ListFiles(line, builtins.c)
				})),
				readline.PcItem("type"),
				readline.PcItem("complete"),
				readline.PcItem("echo"),
				readline.PcItemDynamic(func(line string) []string {
					return path.ListExecutables(line, builtins.c)
				},
					readline.PcItemDynamic(func(line string) []string {
						return path.ListProgrammable(line, builtins.c)
					}),
					readline.PcItemDynamic(func(line string) []string {
						return path.ListFiles(line, builtins.c)
					}),
				),
			),
		},
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

type Builtins struct {
	c *builtin.CompleteCommand
}

func Start() error {
	builtins := &Builtins{
		c: builtin.NewCompleteCommand(),
	}
	reader, err := newReader(builtins)
	if err != nil {
		return err
	}
	defer reader.Close()

	for {
		line, err := reader.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		tokens := parser.Tokenize(strings.TrimSpace(line))
		cmdPipeline, err := parser.Parse(tokens)
		if err != nil {
			if !errors.Is(err, parser.ErrEmptyInput) {
				fmt.Println(err)
			}
			continue
		}

		if len(cmdPipeline.Commands) == 1 && cmdPipeline.Commands[0].Redirects == nil {
			cmd := cmdPipeline.Commands[0]
			builtinFunc := builtinRouter(cmd.Name, cmd.Args, builtins)

			if builtinFunc != nil {
				builtinFunc()
				continue
			}

			if _, err := path.FindExecutable(cmd.Name); err == nil {
				runner.Execute(cmd.Name, cmd.Args...)
			} else {
				fmt.Printf("%s: command not found\n", cmd.Name)
			}
		} else {
			if err := runner.ExecutePipeline(cmdPipeline); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func builtinRouter(command string, rest []string, builtins *Builtins) func() {
	switch command {
	case "type":
		return func() { builtin.Type(rest[0]) }
	case "echo":
		return func() { builtin.Echo(rest) }
	case "exit":
		return func() { builtin.Exit() }
	case "pwd":
		return func() { builtin.Pwd() }
	case "cd":
		return func() {
			err := builtin.Cd(rest[0])
			if err != nil {
				fmt.Println(err)
			}
		}
	case "complete":
		return func() {
			builtins.c.Complete(rest)
		}
	default:
		return nil
	}
}
