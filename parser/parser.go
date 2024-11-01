package parser

import (
	"fmt"
	"strconv"

	"github.com/thacuber2a03/cymbal/ast"
	"github.com/thacuber2a03/cymbal/lexer"
)

type ParseError struct {
	Token   lexer.Token
	Message string
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("Error at line %d: %s", pe.Token.Line, pe.Message)
}

type (
	precedence int // enum
	infixFn    func(*Parser, ast.Expression) ast.Expression
	prefixFn   func(*Parser) ast.Expression
	parseRule  struct {
		prefix prefixFn
		infix  infixFn
		prec   precedence
	}
)

const (
	pNone precedence = iota
	pTerm
	pFactor
	pUnary
	pPrimary
)

type Parser struct {
	*lexer.Lexer
	curTok, nextTok lexer.Token

	Errors    []ParseError
	panicMode bool

	parseRules map[lexer.TokenType]parseRule
}

func (p *Parser) error(tok lexer.Token, msg string) {
	// error engineering taken from clox
	// yeah, the language from Crafting Interpreters

	if p.panicMode {
		return
	}
	p.panicMode = true

	p.Errors = append(p.Errors, ParseError{
		Token:   tok,
		Message: msg,
	})
}

func (p *Parser) advance() (res bool) {
	p.curTok = p.nextTok
	res = true
	for {
		p.nextTok = p.Lexer.Next()
		if p.nextTok.Type != lexer.TT_ERROR {
			return
		}
		res = false
		p.error(p.nextTok, p.nextTok.Lexeme)
	}
}

func (p *Parser) check(tt lexer.TokenType) bool { return p.curTok.Type == tt }

func (p *Parser) atEnd() bool { return p.curTok.Type == lexer.TT_EOF }

func (p *Parser) match(tt lexer.TokenType) bool {
	if p.check(tt) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(tt lexer.TokenType) bool {
	if !p.match(tt) {
		msg := fmt.Sprintf("expected %v, but got %v", tt, p.curTok.Type)
		p.error(p.curTok, msg)
		return false
	}
	return true
}

func (p *Parser) leftAssoc(left ast.Expression) ast.Expression {
	t := p.curTok
	prec := precedence(int(p.parseRules[t.Type].prec) + 1)
	p.advance()
	return &ast.Binary{Left: left, Operator: t, Right: p.parsePrecedence(prec)}
}

func (p *Parser) literal() ast.Expression {
	switch p.curTok.Type {
	case lexer.TT_INT:
		val, err := strconv.ParseInt(p.curTok.Lexeme, 0, 16)
		if err != nil {
			p.error(p.curTok, "integer overflow/underflow")
			return nil
		}
		return &ast.Literal{Value: int16(val)}
	case lexer.TT_CHAR:
		// TODO(thacuber2a03): escapes
		return &ast.Literal{Value: int16(p.curTok.Lexeme[1])}
	default:
		panic("unreachable")
	}
}

func (p *Parser) grouping() (res ast.Expression) {
	p.advance()
	res = p.parsePrecedence(pNone)
	if !p.check(lexer.TT_RPAREN) {
		// TODO(thacuber2a03): "...to close off '(' at <pos>..."
		p.error(p.curTok, "expected ')' after expression")
		return nil
	}
	p.advance()
	return
}

// Parses an expression of precedence [prec].
// Expects the bound functions to not consume its final input.
func (p *Parser) parsePrecedence(prec precedence) (res ast.Expression) {
	rule, ok := p.parseRules[p.curTok.Type]
	if !ok || rule.prefix == nil {
		p.error(p.curTok, "expected expression")
		return
	}

	if res = rule.prefix(p); res == nil {
		return
	}

	for {
		rule, ok := p.parseRules[p.nextTok.Type]
		if !ok || prec > rule.prec {
			break
		}

		if !p.advance() {
			return nil
		}
		
		if res = rule.infix(p, res); res == nil {
			return
		}
	}

	return
}

func (p *Parser) expression() ast.Expression { return p.parsePrecedence(pNone) }

func (p *Parser) deoStmt() *ast.DEOStatement {
	deo := &ast.DEOStatement{}

	if deo.Port = p.expression(); deo.Port == nil {
		return nil
	}

	if !p.expect(lexer.TT_COMMA) {
		return nil
	}

	if deo.Value = p.expression(); deo.Value == nil {
		return nil
	}

	return deo
}

func (p *Parser) statement() ast.Statement {
	if p.match(lexer.TT_DEO) {
		return p.deoStmt()
	}
	return nil
}

func (p *Parser) block() *ast.Block {
	b := &ast.Block{}

	for !(p.atEnd() || p.check(lexer.TT_RBRACE)) {
		s := p.statement()
		if s != nil {
			b.Statements = append(b.Statements, s)
		}
	}

	if !p.expect(lexer.TT_RBRACE) {
		return nil
	}

	return b
}

func (p *Parser) mainDecl() *ast.MainDecl {
	if !p.expect(lexer.TT_LBRACE) {
		return nil
	}
	return (*ast.MainDecl)(p.block())
}

func (p *Parser) declaration() ast.Declaration {
	if p.match(lexer.TT_MAIN) {
		return p.mainDecl()
	}
	return nil
}

func (p *Parser) program() *ast.Program {
	prog := &ast.Program{}

	for !p.check(lexer.TT_EOF) {
		d := p.declaration()
		if d != nil {
			prog.Declarations = append(prog.Declarations, d)
		}
	}

	return prog
}

// Parses the program and returns an AST representation of it.
// Returns nil if there has been any errors.
func (p *Parser) Parse() *ast.Program {
	if !p.advance() {
		return nil
	}
	if !p.advance() {
		return nil
	}

	prog := p.program()
	if !p.expect(lexer.TT_EOF) {
		return nil
	}
	return prog
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{
		Lexer: l,
		parseRules: map[lexer.TokenType]parseRule{
			lexer.TT_INT:    {prefix: (*Parser).literal},
			lexer.TT_CHAR:   {prefix: (*Parser).literal},
			lexer.TT_LPAREN: {prefix: (*Parser).grouping},
			lexer.TT_PLUS:   {infix: (*Parser).leftAssoc, prec: pTerm},
			lexer.TT_MINUS:  {infix: (*Parser).leftAssoc, prec: pTerm},
			lexer.TT_STAR:   {infix: (*Parser).leftAssoc, prec: pFactor},
			lexer.TT_SLASH:  {infix: (*Parser).leftAssoc, prec: pFactor},
		},
	}
}
