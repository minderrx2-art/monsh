package completer

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/chzyer/readline"
)

type ShellCompleter struct {
	ReadlineCompleter readline.PrefixCompleter
	lastPrefix        string
	tabPressed        bool
}

func (c *ShellCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	typed := string(line[:pos])
	rawMatches, offset := c.ReadlineCompleter.Do(line, pos)

	matches := make(map[string]struct{})
	for _, match := range rawMatches {
		matches[string(match)] = struct{}{}
	}
	uniqueMatches := slices.Sorted(maps.Keys(matches))

	if typed != c.lastPrefix {
		c.lastPrefix = typed
		c.tabPressed = false
	}

	if len(uniqueMatches) == 0 {
		fmt.Print("\x07")
		return nil, 0
	}

	if len(uniqueMatches) == 1 {
		c.tabPressed = false
		m := uniqueMatches[0]
		// Keep trailing '/' so the next tab can complete inside the directory.
		if strings.HasSuffix(m, "/ ") {
			m = strings.TrimSuffix(m, " ")
		}
		return [][]rune{[]rune(m)}, offset
	}

	matchBase := typed
	if offset > 0 && offset <= len(typed) {
		matchBase = typed[:offset]
	}
	fullNames := make([]string, 0, len(uniqueMatches))
	for _, suffix := range uniqueMatches {
		fullNames = append(fullNames, matchBase+strings.TrimSpace(suffix))
	}
	lcp := longestCommonPrefix(fullNames)
	if len(lcp) > len(typed) {
		c.tabPressed = false
		return [][]rune{[]rune(lcp[len(typed):])}, offset
	}

	// Complete as far as all suffixes agree.
	suffixLCP := longestCommonPrefix(uniqueMatches)
	if suffixLCP != "" {
		c.tabPressed = false
		return [][]rune{[]rune(suffixLCP)}, offset
	}

	if !c.tabPressed {
		fmt.Print("\x07")
		c.tabPressed = true
		return nil, 0
	}

	// Second tab: list full candidate names, then redraw the prompt line.
	displayPrefix := typed
	if offset > 0 && offset <= len(typed) {
		displayPrefix = typed[len(typed)-offset:]
	}
	names := make([]string, 0, len(uniqueMatches))
	for _, suffix := range uniqueMatches {
		names = append(names, displayPrefix+strings.TrimSpace(suffix))
	}
	slices.Sort(names)
	fmt.Printf("\n%s\n", strings.Join(names, "  "))
	fmt.Printf("$ %s", typed)
	c.tabPressed = false
	return nil, 0
}

func longestCommonPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}
	shortest := names[0]
	for _, s := range names[1:] {
		if len(s) < len(shortest) {
			shortest = s
		}
	}
	for i := 0; i < len(shortest); i++ {
		for _, s := range names {
			if s[i] != shortest[i] {
				return shortest[:i]
			}
		}
	}
	return shortest
}
