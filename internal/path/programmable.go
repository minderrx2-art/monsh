package path

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func completerArgv(line string) (argv1, argv2, argv3 string) {
	trimmed := strings.TrimLeft(line, " \t")
	cmd := getFirstWord(trimmed)
	word := wordBeingCompleted(line)
	rest := strings.TrimLeft(trimmed[len(cmd):], " \t")

	argv1 = cmd
	argv2 = word
	argv3 = cmd
	if word != "" {
		if i := strings.LastIndex(rest, " "); i >= 0 {
			argv3 = rest[:i]
		}
	}
	return argv1, argv2, argv3
}

func runProgrammableCompleter(completerPath, line string) []string {
	trimmed := strings.TrimLeft(line, " \t")
	argv1, argv2, argv3 := completerArgv(trimmed)

	cmd := exec.Command(completerPath, argv1, argv2, argv3)
	cmd.Env = append(os.Environ(),
		"COMP_LINE="+trimmed,
		"COMP_POINT="+strconv.Itoa(len(trimmed)),
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil
	}

	if err := cmd.Start(); err != nil {
		return nil
	}

	stdoutLines := readAvailableLines(stdoutPipe)
	stderrLines := readAvailableLines(stderrPipe)

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()

	if len(stdoutLines) > 0 {
		return stdoutLines
	}

	for i, l := range stderrLines {
		if i > 0 {
			fmt.Println(l)
		}
	}
	if len(stderrLines) > 0 {
		return []string{stderrLines[0]}
	}
	return nil
}

func readAvailableLines(r io.Reader) []string {
	deadline, ok := r.(interface{ SetReadDeadline(time.Time) error })
	if ok {
		_ = deadline.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	}

	var lines []string
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if n := len(line); n > 0 {
			lines = append(lines, strings.TrimRight(line, "\r\n"))
		}
		if err != nil {
			break
		}
	}
	return lines
}
