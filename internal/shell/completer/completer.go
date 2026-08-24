package completer

import (
	"fmt"

	"github.com/chzyer/readline"
)

type ShellCompleter struct {
	ReadlineCompleter readline.PrefixCompleter
	lastPrefix        string
	tabPressed        bool
	matches           [][]rune
	offset            int
}

func (c *ShellCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	c.matches, c.offset = c.ReadlineCompleter.Do(line, pos)
	c.handleBell()
	c.handleSmallInput(line)

	return c.matches, c.offset
}

func (c *ShellCompleter) handleBell() {
	if len(c.matches) > 0 {
		return
	}
	if c.tabPressed {
		return
	}
	fmt.Print("\x07")
}

func (c *ShellCompleter) handleSmallInput(line []rune) {
	if len(line) == 1 {
		fmt.Printf("\nauto_complete: require more symbols to auto complete '%s' ", string(line))
	}
}
