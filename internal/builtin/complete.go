package builtin

import (
	"fmt"
	"sort"
)

type CompleteCommand struct {
	savedCompletions map[string]string
}

func (c *CompleteCommand) Register(name string, completerPath string) {
	c.savedCompletions[name] = completerPath
}

func (c *CompleteCommand) Get(name string) (string, bool) {
	saved, ok := c.savedCompletions[name]
	return saved, ok
}

func NewCompleteCommand() *CompleteCommand {
	return &CompleteCommand{
		savedCompletions: make(map[string]string),
	}
}

func (c *CompleteCommand) Complete(args []string) {
	if len(args) == 0 {
		names := mapsKeys(c.savedCompletions)
		sort.Strings(names)
		for _, name := range names {
			path := c.savedCompletions[name]
			fmt.Printf("complete -C '%s' %s\n", path, name)
		}
		return
	}

	switch args[0] {
	case "-p":
		path, ok := c.Get(args[1])
		if !ok {
			fmt.Printf("complete: %s: no completion specification\n", args[1])
			return
		}
		fmt.Printf("complete -C '%s' %s\n", path, args[1])
	case "-C":
		c.Register(args[2], args[1])
	default:
		fmt.Printf("complete: %s: no completion specification\n", args[0])
	}
}

func mapsKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
