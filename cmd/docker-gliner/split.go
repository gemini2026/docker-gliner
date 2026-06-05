package main

import "strings"

// splitCommand splits a command string into argv with minimal shell-like
// quoting: whitespace separates words, single and double quotes group, and a
// backslash escapes the next character. It is not a full shell parser — it
// covers the `command:` option for cases a plain field split would mangle.
func splitCommand(s string) []string {
	var args []string
	var cur strings.Builder
	inWord := false
	var quote rune // 0, '\'' or '"'
	escaped := false

	flush := func() {
		if inWord {
			args = append(args, cur.String())
			cur.Reset()
			inWord = false
		}
	}

	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			inWord = true
			escaped = false
		case r == '\\' && quote != '\'':
			escaped = true
			inWord = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inWord = true
		case r == ' ' || r == '\t' || r == '\n':
			flush()
		default:
			cur.WriteRune(r)
			inWord = true
		}
	}
	flush()
	return args
}
