package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	source string
	index  int
	line   uint
}

func New(source string) *Lexer {
	return &Lexer{
		source: source,
		line:   1,
	}
}

func (l *Lexer) atEnd() bool { return l.index >= len(l.source) }

func (l *Lexer) peekChar() byte {
	if l.atEnd() {
		return 0
	}
	return l.source[l.index]
}

func (l *Lexer) nextChar() byte {
	ch := l.peekChar()
	l.index++
	if ch == '\n' {
		l.line++
	}
	return ch
}

func (l *Lexer) token(tt TokenType, lexeme string) Token {
	return Token{Type: tt, Lexeme: lexeme, Line: l.line}
}

func (l *Lexer) error(msg string, args ...any) Token {
	return l.token(TT_ERROR, fmt.Sprintf(msg, args...))
}

func isValidIdent(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

var keywords = map[string]TokenType{
	"fn":   TT_FN,
	"main": TT_MAIN,
	"deo":  TT_DEO,
}

func (l *Lexer) ident(c byte) Token {
	var id strings.Builder
	id.WriteByte(c)

	for !l.atEnd() && isValidIdent(l.peekChar()) ||
		unicode.IsDigit(rune(l.peekChar())) {
		id.WriteByte(l.nextChar())
	}

	if id.Len() == 0 {
		panic("(*Lexer).ident was called but no identifier was scanned")
	}

	str := id.String()
	if k, ok := keywords[str]; ok {
		return l.token(k, str)
	}
	return l.token(TT_IDENT, str)
}

var validBases = map[byte]struct {
	fn   func(r byte) bool
	name string
}{
	'x': {
		fn:   func(ch byte) bool { return unicode.Is(unicode.ASCII_Hex_Digit, rune(ch)) },
		name: "hexadecimal",
	},
	'b': {
		fn:   func(ch byte) bool { return ch == '0' || ch == '1' },
		name: "binary",
	},
}

func (l *Lexer) number(c byte) Token {
	var num strings.Builder
	num.WriteByte(c)

	for !l.atEnd() && unicode.IsDigit(rune(l.peekChar())) {
		num.WriteByte(l.nextChar())
	}

	if num.Len() == 0 {
		panic("(*Lexer).number was called but no integer literal was scanned")
	}

	if b, ok := validBases[l.peekChar()]; !l.atEnd() && ok {
		num.WriteByte(l.nextChar())

		var res strings.Builder

		for !l.atEnd() && b.fn(l.peekChar()) {
			res.WriteByte(l.nextChar())
		}

		if res.Len() == 0 {
			return l.error("no digits in %s literal", b.name)
		}

		num.WriteString(res.String())
	}

	return l.token(TT_INT, num.String())
}

func (l *Lexer) charLit(c byte) Token {
	var lit strings.Builder
	lit.WriteByte(c)
	if l.atEnd() {
		return l.error("unterminated character literal")
	}

	if c := l.nextChar(); c == '\\' {
		// TODO(thacuber2a03): escapes
	} else {
		lit.WriteByte(c)
	}

	if l.atEnd() {
		return l.error("unterminated character literal")
	}
	lit.WriteByte(l.nextChar())
	return l.token(TT_CHAR, lit.String())
}

func (l *Lexer) skipWhitespace() {
out:
	for {
		switch l.peekChar() {
		// comment
		case '#':
			for !l.atEnd() && l.peekChar() != '\n' {
				l.nextChar()
			}

		case ' ', '\t', '\n', '\r', '\v', '\f':
			l.nextChar()

		default:
			break out
		}
	}
}

func (l *Lexer) Next() Token {
	l.skipWhitespace()

	if l.atEnd() {
		return l.token(TT_EOF, "<eof>")
	}

	c := l.nextChar()

	switch c {
	case '{':
		return l.token(TT_LBRACE, "{")
	case '}':
		return l.token(TT_RBRACE, "}")
	case '(':
		return l.token(TT_LPAREN, "(")
	case ')':
		return l.token(TT_RPAREN, ")")
	case ',':
		return l.token(TT_COMMA, ",")
	case '+':
		return l.token(TT_PLUS, "+")
	case '-':
		return l.token(TT_MINUS, "-")
	case '*':
		return l.token(TT_STAR, "*")
	case '/':
		return l.token(TT_SLASH, "/")
	case '\'':
		return l.charLit(c)
	}

	if isValidIdent(c) {
		return l.ident(c)
	} else if unicode.IsDigit(rune(c)) {
		return l.number(c)
	}

	return l.error("unrecognized character '%q'", c)
}
