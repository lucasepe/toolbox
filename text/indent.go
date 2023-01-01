package text

import (
	"strings"
)

// indents a block of text with a specified indent string
func Indent(text, indent string) string {
	if len(text) == 0 {
		return text
	}

	var sb strings.Builder

	var parts []string
	if text[len(text)-1:] == "\n" {
		parts = strings.Split(text[:len(text)-1], "\n")
	} else {
		parts = strings.Split(strings.TrimRight(text, "\n"), "\n")
	}

	for _, j := range parts {
		sb.WriteString(indent)
		sb.WriteString(j)
		sb.WriteString("\n")
	}

	return sb.String()[:sb.Len()-1]
}
