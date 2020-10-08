package typescriptify

import "strings"

func indentLines(str string, i int) string {
	lines := strings.Split(str, "\n")
	for n := range lines {
		lines[n] = strings.Repeat("\t", i) + lines[n]
	}
	return strings.Join(lines, "\n")
}
