package path

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type CompleterRegistry interface {
	Get(name string) (string, bool)
}

var shellBuiltins = map[string]struct{}{
	"cd": {}, "echo": {}, "exit": {}, "pwd": {}, "type": {},
}

func scanPath() []string {
	PATH := os.Getenv("PATH")
	paths := strings.Split(PATH, ":")
	seen := make(map[string]struct{})
	matches := []string{}
	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}
			if info.Mode().Perm()&0111 != 0 {
				name := file.Name()
				if _, ok := shellBuiltins[name]; ok {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				matches = append(matches, name)
			}
		}
	}
	return matches
}

func FindExecutable(binary string) (string, error) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("Executable not found")
	}
	return path, nil
}

func getFirstWord(line string) string {
	line = strings.TrimLeft(line, " \t")
	if line == "" {
		return ""
	}
	if i := strings.IndexAny(line, " \t"); i >= 0 {
		return line[:i]
	}
	return line
}

// "ls bee" -> returns "bee"
func wordBeingCompleted(line string) string {
	if i := strings.LastIndex(line, " "); i >= 0 {
		return line[i+1:]
	}
	return ""
}

func completingArgs(line string) bool {
	return strings.Contains(strings.TrimLeft(line, " \t"), " ")
}

func ListExecutables(line string, registry CompleterRegistry) []string {
	word := getFirstWord(line)
	if word == "" {
		return nil
	}
	if completingArgs(line) {
		if _, builtin := shellBuiltins[word]; builtin {
			return nil
		}
		if registry != nil {
			if _, ok := registry.Get(word); ok {
				return []string{word}
			}
		}
		if _, err := exec.LookPath(word); err == nil {
			return []string{word}
		}
		return nil
	}

	if _, builtin := shellBuiltins[word]; builtin {
		return nil
	}

	matches := slices.DeleteFunc(scanPath(), func(executable string) bool {
		return !strings.HasPrefix(executable, word)
	})
	if len(matches) == 0 {
		return nil
	}
	return matches
}

func argPrefixBeforeLastWord(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	cmd := getFirstWord(trimmed)
	rest := strings.TrimLeft(trimmed[len(cmd):], " \t")
	if i := strings.LastIndex(rest, " "); i >= 0 {
		return rest[:i+1]
	}
	return ""
}

func ListFiles(line string, registry CompleterRegistry) []string {
	if !completingArgs(line) {
		return nil
	}

	if registry != nil {
		if _, ok := registry.Get(getFirstWord(line)); ok {
			return nil
		}
	}

	word := wordBeingCompleted(line)
	argPrefix := argPrefixBeforeLastWord(line)

	dir := "."
	prefix := word

	if i := strings.LastIndex(word, "/"); i >= 0 {
		dir = word[:i+1]
		prefix = word[i+1:]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	files, err := os.ReadDir(filepath.Join(cwd, dir))
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(files))
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), prefix) {
			continue
		}
		name := file.Name()
		if dir != "." {
			name = dir + name
		}
		if file.IsDir() {
			name += "/"
		}
		names = append(names, argPrefix+name)
	}
	return names
}

func ListProgrammable(line string, registry CompleterRegistry) []string {
	if !completingArgs(line) {
		return nil
	}

	completerPath, ok := registry.Get(getFirstWord(line))
	if !ok {
		return nil
	}

	candidates := runProgrammableCompleter(completerPath, line)
	if len(candidates) == 0 {
		return nil
	}

	word := wordBeingCompleted(line)
	argPrefix := argPrefixBeforeLastWord(line)

	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimRight(candidate, "\r")
		if candidate == "" {
			continue
		}
		if word != "" && !strings.HasPrefix(candidate, word) {
			continue
		}
		names = append(names, argPrefix+candidate)
	}
	return names
}
