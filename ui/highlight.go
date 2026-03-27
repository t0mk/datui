package ui

import (
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

func highlightYAMLLine(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	// Document markers
	if trimmed == "---" || trimmed == "..." {
		return "[darkgray::b]" + tview.Escape(line) + "[-:-:-]"
	}
	// Empty
	if trimmed == "" {
		return ""
	}
	// Comment
	if strings.HasPrefix(trimmed, "#") {
		return "[gray]" + tview.Escape(line) + "[-]"
	}

	// Strip optional list marker "- " from the front
	listPrefix := ""
	rest := trimmed
	if strings.HasPrefix(rest, "- ") {
		listPrefix = "[navy]-[-] "
		rest = rest[2:]
	} else if rest == "-" {
		return indent + "[navy]-[-]"
	}

	// Find key:value split
	colonIdx := findKeyColon(rest)
	if colonIdx >= 0 {
		rawKey := rest[:colonIdx]
		after := rest[colonIdx+1:] // everything after ":"

		// Separate optional space after colon
		spaceEnd := 0
		for spaceEnd < len(after) && after[spaceEnd] == ' ' {
			spaceEnd++
		}
		sep := ":" + after[:spaceEnd]
		val := after[spaceEnd:]

		return indent + listPrefix +
			"[navy::b]" + tview.Escape(rawKey) + "[-:-:-]" +
			"[darkgray]" + tview.Escape(sep) + "[-]" +
			colorizeValue(val)
	}

	// List item value (no key)
	if listPrefix != "" {
		return indent + listPrefix + colorizeValue(rest)
	}

	// Plain continuation (block scalar body, etc.)
	return tview.Escape(line)
}

// findKeyColon returns the index of the YAML key-separator colon, or -1.
func findKeyColon(s string) int {
	inSingle, inDouble := false, false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case ':':
			if inSingle || inDouble {
				continue
			}
			// Skip "://" (URLs)
			if i+2 < len(s) && s[i+1] == '/' && s[i+2] == '/' {
				continue
			}
			// Valid separator: followed by space, tab, end-of-string
			if i+1 >= len(s) || s[i+1] == ' ' || s[i+1] == '\t' {
				return i
			}
		}
	}
	return -1
}

func colorizeValue(val string) string {
	if val == "" {
		return ""
	}
	t := strings.TrimSpace(val)
	if t == "" {
		return tview.Escape(val)
	}
	// Block scalar indicators
	if t == "|" || t == ">" || t == "|-" || t == ">-" || t == "|+" || t == ">+" {
		return "[darkgray]" + tview.Escape(val) + "[-]"
	}
	// Quoted string
	if len(t) >= 2 &&
		((t[0] == '"' && t[len(t)-1] == '"') || (t[0] == '\'' && t[len(t)-1] == '\'')) {
		return "[teal]" + tview.Escape(val) + "[-]"
	}
	// Boolean
	switch strings.ToLower(t) {
	case "true", "false", "yes", "no", "on", "off":
		return "[purple]" + tview.Escape(val) + "[-]"
	}
	// Null
	switch strings.ToLower(t) {
	case "null", "~":
		return "[darkgray]" + tview.Escape(val) + "[-]"
	}
	// Number
	if _, err := strconv.ParseInt(t, 10, 64); err == nil {
		return "[maroon]" + tview.Escape(val) + "[-]"
	}
	if _, err := strconv.ParseFloat(t, 64); err == nil {
		return "[maroon]" + tview.Escape(val) + "[-]"
	}
	// Default: string value
	return "[darkgreen]" + tview.Escape(val) + "[-]"
}
