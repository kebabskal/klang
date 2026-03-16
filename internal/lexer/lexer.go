package lexer

import "fmt"

// LexError represents a lexer error with position info.
type LexError struct {
	Line    int
	Col     int
	Message string
}

func (e LexError) Error() string {
	return fmt.Sprintf("line %d:%d: %s", e.Line, e.Col, e.Message)
}

type Lexer struct {
	src        []byte
	pos        int
	line       int
	col        int
	tokens     []Token
	errors     []LexError
	atBOL      bool // at beginning of line (for indent tracking)
	parenDepth int  // tracks () and [] nesting; newlines suppressed when > 0
}

func New(src []byte) *Lexer {
	return &Lexer{
		src:   src,
		pos:   0,
		line:  1,
		col:   1,
		atBOL: true,
	}
}

// Source returns the raw source bytes (for error reporting).
func (l *Lexer) Source() []byte {
	return l.src
}

// Errors returns any errors encountered during lexing.
func (l *Lexer) Errors() []LexError {
	return l.errors
}

func (l *Lexer) addError(msg string) {
	l.errors = append(l.errors, LexError{Line: l.line, Col: l.col, Message: msg})
}

func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.src) {
		l.skipComment()
		if l.pos >= len(l.src) {
			break
		}

		ch := l.src[l.pos]

		// Handle newlines
		if ch == '\n' {
			l.pos++
			l.line++
			l.col = 1
			if l.parenDepth > 0 {
				// Inside parens/brackets: suppress newline tokens
				continue
			}
			l.emit(TOKEN_NEWLINE, "\n")
			l.atBOL = true
			continue
		}
		if ch == '\r' {
			l.pos++
			if l.pos < len(l.src) && l.src[l.pos] == '\n' {
				l.pos++
			}
			l.line++
			l.col = 1
			if l.parenDepth > 0 {
				continue
			}
			l.emit(TOKEN_NEWLINE, "\n")
			l.atBOL = true
			continue
		}

		// Measure indentation at beginning of line (skip inside parens/brackets)
		if l.atBOL && l.parenDepth == 0 {
			indent := 0
			for l.pos < len(l.src) && (l.src[l.pos] == '\t' || l.src[l.pos] == ' ') {
				if l.src[l.pos] == '\t' {
					indent++
				} else {
					indent++ // treat each space as 1 indent unit; tabs also 1
				}
				l.pos++
				l.col++
			}
			l.atBOL = false
			// Skip if line is empty or comment
			if l.pos >= len(l.src) || l.src[l.pos] == '\n' || l.src[l.pos] == '\r' || l.src[l.pos] == '#' {
				// store indent for the next real token, but don't emit indent token for blank lines
				continue
			}
			if indent > 0 {
				l.emit(TOKEN_INDENT, "")
				l.tokens[len(l.tokens)-1].Indent = indent
			}
			continue
		}

		// Skip whitespace (non-newline)
		if ch == ' ' || ch == '\t' {
			l.pos++
			l.col++
			continue
		}

		// String interpolation $"..."
		if ch == '$' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '"' {
			l.readInterpString()
			continue
		}

		// String literal
		if ch == '"' {
			l.readString()
			continue
		}

		// Number
		if isDigit(ch) {
			l.readNumber()
			continue
		}

		// Identifier / keyword
		if isIdentStart(ch) {
			l.readIdent()
			continue
		}

		// Inline C: @c { raw code }
		if ch == '@' {
			l.readInlineC()
			continue
		}

		// Multi-char operators
		if ch == ':' && l.peek(1) == '=' {
			l.emit(TOKEN_COLON_EQ, ":=")
			l.advance(2)
			continue
		}
		if ch == '=' && l.peek(1) == '>' {
			l.emit(TOKEN_ARROW, "=>")
			l.advance(2)
			continue
		}
		if ch == '=' && l.peek(1) == '=' {
			l.emit(TOKEN_EQEQ, "==")
			l.advance(2)
			continue
		}
		if ch == '!' && l.peek(1) == '=' {
			l.emit(TOKEN_NEQ, "!=")
			l.advance(2)
			continue
		}
		if ch == '<' && l.peek(1) == '=' {
			l.emit(TOKEN_LTEQ, "<=")
			l.advance(2)
			continue
		}
		if ch == '>' && l.peek(1) == '=' {
			l.emit(TOKEN_GTEQ, ">=")
			l.advance(2)
			continue
		}
		if ch == '+' && l.peek(1) == '=' {
			l.emit(TOKEN_PLUS_EQ, "+=")
			l.advance(2)
			continue
		}
		if ch == '-' && l.peek(1) == '=' {
			l.emit(TOKEN_MINUS_EQ, "-=")
			l.advance(2)
			continue
		}
		if ch == '*' && l.peek(1) == '=' {
			l.emit(TOKEN_STAR_EQ, "*=")
			l.advance(2)
			continue
		}
		if ch == '/' && l.peek(1) == '=' {
			l.emit(TOKEN_SLASH_EQ, "/=")
			l.advance(2)
			continue
		}
		if ch == '%' && l.peek(1) == '=' {
			l.emit(TOKEN_PERCENT_EQ, "%=")
			l.advance(2)
			continue
		}
		if ch == '?' && l.peek(1) == '.' {
			l.emit(TOKEN_QUESTION_DOT, "?.")
			l.advance(2)
			continue
		}
		if ch == '?' && l.peek(1) == '?' {
			l.emit(TOKEN_QUESTION_QUESTION, "??")
			l.advance(2)
			continue
		}
		if ch == '&' && l.peek(1) == '&' {
			l.emit(TOKEN_AMPAMP, "&&")
			l.advance(2)
			continue
		}
		if ch == '|' && l.peek(1) == '|' {
			l.emit(TOKEN_PIPEPIPE, "||")
			l.advance(2)
			continue
		}
		if ch == '.' && l.peek(1) == '.' && l.peek(2) != '.' {
			l.emit(TOKEN_DOTDOT, "..")
			l.advance(2)
			continue
		}
		if ch == '.' && l.peek(1) == '.' && l.peek(2) == '.' {
			l.emit(TOKEN_ELLIPSIS, "...")
			l.advance(3)
			continue
		}

		// Single-char operators
		switch ch {
		case ':':
			l.emit(TOKEN_COLON, ":")
		case '=':
			l.emit(TOKEN_EQ, "=")
		case '.':
			l.emit(TOKEN_DOT, ".")
		case ',':
			l.emit(TOKEN_COMMA, ",")
		case '+':
			l.emit(TOKEN_PLUS, "+")
		case '-':
			l.emit(TOKEN_MINUS, "-")
		case '*':
			l.emit(TOKEN_STAR, "*")
		case '/':
			l.emit(TOKEN_SLASH, "/")
		case '%':
			l.emit(TOKEN_PERCENT, "%")
		case '<':
			l.emit(TOKEN_LT, "<")
		case '>':
			l.emit(TOKEN_GT, ">")
		case '|':
			l.emit(TOKEN_PIPE, "|")
		case '!':
			l.emit(TOKEN_BANG, "!")
		case '(':
			l.parenDepth++
			l.emit(TOKEN_LPAREN, "(")
		case ')':
			if l.parenDepth > 0 {
				l.parenDepth--
			}
			l.emit(TOKEN_RPAREN, ")")
		case '{':
			l.emit(TOKEN_LBRACE, "{")
		case '}':
			l.emit(TOKEN_RBRACE, "}")
		case '[':
			l.parenDepth++
			l.emit(TOKEN_LBRACKET, "[")
		case ']':
			if l.parenDepth > 0 {
				l.parenDepth--
			}
			l.emit(TOKEN_RBRACKET, "]")
		case '#':
			l.skipComment()
			continue
		default:
			l.addError(fmt.Sprintf("unexpected character '%c'", ch))
		}
		l.pos++
		l.col++
	}

	l.emit(TOKEN_EOF, "")
	return l.tokens
}

func (l *Lexer) emit(typ TokenType, val string) {
	l.tokens = append(l.tokens, Token{
		Type:  typ,
		Value: val,
		Line:  l.line,
		Col:   l.col,
	})
}

func (l *Lexer) peek(offset int) byte {
	idx := l.pos + offset
	if idx >= len(l.src) {
		return 0
	}
	return l.src[idx]
}

func (l *Lexer) advance(n int) {
	l.pos += n
	l.col += n
}

func (l *Lexer) skipComment() {
	if l.pos < len(l.src) && l.src[l.pos] == '#' {
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.pos++
		}
	}
}

func (l *Lexer) readString() {
	l.pos++ // skip opening "
	l.col++
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '"' && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' {
			l.pos++ // skip escape
		}
		l.pos++
	}
	val := string(l.src[start:l.pos])
	if l.pos < len(l.src) && l.src[l.pos] == '"' {
		l.pos++ // skip closing "
	} else {
		l.addError("unterminated string literal")
	}
	l.emit(TOKEN_STRING_LIT, val)
	l.col = l.col + len(val) + 1
}

func (l *Lexer) readInterpString() {
	// $"..." — for now, store as INTERP_STRING with the full content between quotes
	l.pos += 2 // skip $"
	l.col += 2
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '"' && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' {
			l.pos++
		}
		l.pos++
	}
	val := string(l.src[start:l.pos])
	if l.pos < len(l.src) && l.src[l.pos] == '"' {
		l.pos++
	} else {
		l.addError("unterminated interpolated string literal")
	}
	l.emit(TOKEN_INTERP_STRING, val)
	l.col = l.col + len(val) + 1
}

func (l *Lexer) readNumber() {
	start := l.pos
	isFloat := false
	for l.pos < len(l.src) && (isDigit(l.src[l.pos]) || l.src[l.pos] == '.') {
		if l.src[l.pos] == '.' {
			// check if it's a decimal dot vs member access
			if l.pos+1 < len(l.src) && isDigit(l.src[l.pos+1]) {
				isFloat = true
			} else {
				break
			}
		}
		l.pos++
	}
	val := string(l.src[start:l.pos])
	if isFloat {
		l.emit(TOKEN_FLOAT_LIT, val)
	} else {
		l.emit(TOKEN_INT_LIT, val)
	}
	l.col += len(val)
}

func (l *Lexer) readIdent() {
	start := l.pos
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.pos++
	}
	val := string(l.src[start:l.pos])
	typ := LookupKeyword(val)
	l.emit(typ, val)
	l.col += len(val)
}

func (l *Lexer) readInlineC() {
	l.pos++ // skip @
	l.col++
	// Expect 'c' identifier
	if l.pos >= len(l.src) || l.src[l.pos] != 'c' {
		l.emit(TOKEN_AT, "@")
		return
	}
	// Check it's just 'c' not 'class' etc
	if l.pos+1 < len(l.src) && isIdentPart(l.src[l.pos+1]) {
		l.emit(TOKEN_AT, "@")
		return
	}
	l.pos++ // skip 'c'
	l.col++
	// Skip whitespace/newlines until '{'
	for l.pos < len(l.src) && (l.src[l.pos] == ' ' || l.src[l.pos] == '\t' || l.src[l.pos] == '\n' || l.src[l.pos] == '\r') {
		if l.src[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
	if l.pos >= len(l.src) || l.src[l.pos] != '{' {
		l.emit(TOKEN_AT, "@")
		return
	}
	l.pos++ // skip '{'
	l.col++
	// Capture everything until matching '}', tracking brace depth
	depth := 1
	start := l.pos
	for l.pos < len(l.src) && depth > 0 {
		if l.src[l.pos] == '{' {
			depth++
		} else if l.src[l.pos] == '}' {
			depth--
			if depth == 0 {
				break
			}
		} else if l.src[l.pos] == '\n' {
			l.line++
			l.col = 0
		}
		l.pos++
		l.col++
	}
	val := string(l.src[start:l.pos])
	if l.pos < len(l.src) && l.src[l.pos] == '}' {
		l.pos++ // skip closing '}'
		l.col++
	}
	l.emit(TOKEN_INLINE_C, val)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}
