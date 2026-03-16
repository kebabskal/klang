package lexer

import "fmt"

type TokenType int

const (
	// Special
	TOKEN_EOF TokenType = iota
	TOKEN_NEWLINE
	TOKEN_INDENT // track indentation level

	// Literals
	TOKEN_IDENT
	TOKEN_INT_LIT
	TOKEN_FLOAT_LIT
	TOKEN_STRING_LIT
	TOKEN_INTERP_STRING // $"..."

	// Operators
	TOKEN_COLON      // :
	TOKEN_COLON_EQ   // :=
	TOKEN_EQ         // =
	TOKEN_EQEQ       // ==
	TOKEN_NEQ        // !=
	TOKEN_LT         // <
	TOKEN_GT         // >
	TOKEN_LTEQ       // <=
	TOKEN_GTEQ       // >=
	TOKEN_PLUS       // +
	TOKEN_MINUS      // -
	TOKEN_STAR       // *
	TOKEN_SLASH      // /
	TOKEN_PERCENT    // %
	TOKEN_PLUS_EQ    // +=
	TOKEN_MINUS_EQ   // -=
	TOKEN_STAR_EQ    // *=
	TOKEN_SLASH_EQ   // /=
	TOKEN_PERCENT_EQ // %=
	TOKEN_DOT        // .
	TOKEN_QUESTION_DOT // ?.
	TOKEN_QUESTION_QUESTION // ??
	TOKEN_ARROW      // =>
	TOKEN_PIPE       // |
	TOKEN_BANG       // !
	TOKEN_DOLLAR     // $
	TOKEN_HASH       // #
	TOKEN_COMMA      // ,
	TOKEN_AMPERSAND  // &
	TOKEN_AMPAMP     // &&
	TOKEN_PIPEPIPE   // ||

	// Delimiters
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]

	// Keywords
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_RETURN
	TOKEN_NOT
	TOKEN_IS
	TOKEN_CLASS
	TOKEN_ENUM
	TOKEN_EVENT
	TOKEN_GET
	TOKEN_SET
	TOKEN_NAMESPACE
	TOKEN_CONSTRUCTOR
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_AND
	TOKEN_OR
	TOKEN_WHILE
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_FN
	TOKEN_WITH
	TOKEN_DOTDOT     // ..
	TOKEN_ELLIPSIS   // ...
	TOKEN_AT         // @
	TOKEN_INLINE_C   // @c { raw C code }
)

var keywords = map[string]TokenType{
	"if":          TOKEN_IF,
	"else":        TOKEN_ELSE,
	"for":         TOKEN_FOR,
	"in":          TOKEN_IN,
	"return":      TOKEN_RETURN,
	"not":         TOKEN_NOT,
	"is":          TOKEN_IS,
	"class":       TOKEN_CLASS,
	"enum":        TOKEN_ENUM,
	"event":       TOKEN_EVENT,
	"get":         TOKEN_GET,
	"set":         TOKEN_SET,
	"namespace":   TOKEN_NAMESPACE,
	"new":         TOKEN_CONSTRUCTOR,
	"this":        TOKEN_THIS,
	"true":        TOKEN_TRUE,
	"false":       TOKEN_FALSE,
	"and":         TOKEN_AND,
	"or":          TOKEN_OR,
	"while":       TOKEN_WHILE,
	"break":       TOKEN_BREAK,
	"continue":    TOKEN_CONTINUE,
	"fn":          TOKEN_FN,
	"with":        TOKEN_WITH,
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
	Indent  int // indentation level (tabs/spaces at start of line)
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%d, %q, %d:%d)", t.Type, t.Value, t.Line, t.Col)
}

func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}
