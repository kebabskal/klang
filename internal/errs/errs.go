package errs

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[31m"
	Yellow    = "\033[33m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BoldRed   = "\033[1;31m"
	BoldCyan  = "\033[1;36m"
	BoldWhite = "\033[1;37m"
)

// colorsEnabled controls whether ANSI codes are emitted.
// Disabled when stderr is not a terminal (e.g. piped).
var colorsEnabled = isTerminal()

func isTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func c(code, text string) string {
	if !colorsEnabled {
		return text
	}
	return code + text + Reset
}

// Kind describes the severity of a diagnostic.
type Kind int

const (
	Error Kind = iota
	Warning
)

// Diagnostic represents a single compiler error or warning.
type Diagnostic struct {
	File    string
	Line    int
	Col     int
	EndCol  int // 0 means underline single char
	Kind    Kind
	Message string
	Source  string // the full source line text
}

// Format renders a diagnostic as a human-readable colored string.
//
//	error: unexpected character '$'
//	  --> example.k:5:12
//	   |
//	 5 |   x := $foo
//	   |        ^
func (d Diagnostic) Format() string {
	var b strings.Builder

	// Header: "error: message" or "warning: message"
	switch d.Kind {
	case Error:
		b.WriteString(c(BoldRed, "error: "))
	case Warning:
		b.WriteString(c(Bold+Yellow, "warning: "))
	}
	b.WriteString(c(BoldWhite, d.Message))
	b.WriteByte('\n')

	// Location: "  --> file:line:col"
	if d.Line > 0 {
		b.WriteString(c(Cyan, "  --> "))
		b.WriteString(c(Dim, fmt.Sprintf("%s:%d:%d", d.File, d.Line, d.Col)))
		b.WriteByte('\n')
	} else {
		b.WriteString(c(Cyan, "  --> "))
		b.WriteString(c(Dim, d.File))
		b.WriteByte('\n')
	}

	// Source line with gutter
	if d.Source != "" {
		lineStr := fmt.Sprintf("%d", d.Line)
		pad := strings.Repeat(" ", len(lineStr))

		// Empty gutter line
		b.WriteString(c(Cyan, fmt.Sprintf(" %s |", pad)))
		b.WriteByte('\n')

		// Source line
		b.WriteString(c(Cyan, fmt.Sprintf(" %s | ", lineStr)))
		b.WriteString(d.Source)
		b.WriteByte('\n')

		// Underline
		col := d.Col
		if col < 1 {
			col = 1
		}
		underlineLen := d.EndCol - d.Col
		if underlineLen < 1 {
			underlineLen = 1
		}
		b.WriteString(c(Cyan, fmt.Sprintf(" %s | ", pad)))
		b.WriteString(strings.Repeat(" ", col-1))
		b.WriteString(c(BoldRed, strings.Repeat("^", underlineLen)))
		b.WriteByte('\n')
	}

	return b.String()
}

// GetSourceLine extracts line number `line` (1-based) from source bytes.
func GetSourceLine(src []byte, line int) string {
	if line < 1 {
		return ""
	}
	current := 1
	i := 0
	// Skip to the start of the requested line
	for i < len(src) && current < line {
		if src[i] == '\n' {
			current++
		}
		i++
	}
	if current != line {
		return ""
	}
	// Find the end of this line
	end := i
	for end < len(src) && src[end] != '\n' && src[end] != '\r' {
		end++
	}
	return string(src[i:end])
}

// FormatSimple formats a simple error without source context.
func FormatSimple(kind Kind, message string) string {
	switch kind {
	case Error:
		return c(BoldRed, "error: ") + c(BoldWhite, message)
	case Warning:
		return c(Bold+Yellow, "warning: ") + c(BoldWhite, message)
	}
	return message
}

// FormatFileError formats an error with file context but no source line.
func FormatFileError(file, message string) string {
	return c(BoldRed, "error: ") + c(BoldWhite, message) + "\n" +
		c(Cyan, "  --> ") + c(Dim, file) + "\n"
}
