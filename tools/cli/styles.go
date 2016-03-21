package cli

import (
	"fmt"
	"strings"
)

// ApplyStyles parses s using markdown-like rules to style CLI text
func ApplyStyles(src string) string {
	parts := strings.Split(" "+src+" ", "**")
	if len(parts) < 3 {
		return src
	}
	out := []string{}
	for i, p := range parts {
		if i%2 == 0 {
			out = append(out, p+"\033[1m")
		} else {
			out = append(out, p+"\033[0m")
		}
	}

	return strings.TrimSuffix(strings.TrimPrefix(strings.Join(out, ""), " "), " ") + "\033[0m"
}

// B makes text bold
func B(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
