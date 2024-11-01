package parser_test

import (
	"fmt"
	"testing"

	"github.com/thacuber2a03/cymbal/ast"
	"github.com/thacuber2a03/cymbal/lexer"
	"github.com/thacuber2a03/cymbal/parser"
)

func parseProgram(t *testing.T, source string) *ast.Program {
	l := lexer.New(source)
	p := parser.New(l)
	prog := p.Parse()

	if prog == nil || len(p.Errors) != 0 {
		t.Errorf("parser failed with %d errors", len(p.Errors))
		for _, e := range p.Errors {
			t.Error(e)
		}
		t.FailNow()
	}

	return prog
}

func parseExpression(t *testing.T, expr string) ast.Expression {
	prog := parseProgram(t, fmt.Sprintf("main { deo 0x18, (%s) }", expr))
	main := prog.Declarations[0].(*ast.MainDecl)
	deo := main.Statements[0].(*ast.DEOStatement)
	return deo.Value
}

func matchLiteral(t *testing.T, e ast.Expression, n string, v int16) bool {
	lit, ok := e.(*ast.Literal)
	if !ok {
		t.Errorf("expected %s to be a Literal, got %T", n, e)
	}

	if lit.Value != v {
		t.Errorf("expected %s to equal %d, got %d instead",
			n, v, lit.Value)
		return false
	}

	return true
}

func TestBasicParsing(t *testing.T) {
	prog := parseProgram(t, `main {
	deo 0x18, 'a'
}`)

	if len(prog.Declarations) != 1 {
		t.Fatalf("expected 1 declaration in the source, got %d",
			len(prog.Declarations))
	}

	main, ok := prog.Declarations[0].(*ast.MainDecl)
	if !ok {
		t.Fatalf("expected single declaration to be a MainDecl, got %T",
			prog.Declarations[0])
	}

	deo, ok := main.Statements[0].(*ast.DEOStatement)
	if !ok {
		t.Fatalf("expected single main statement to be a DEOStatement, got %T",
			main.Statements[0])
	}

	if !matchLiteral(t, deo.Port, "Port", 0x18) {
		return
	}
	if !matchLiteral(t, deo.Value, "Value", int16('a')) {
		return
	}
}

func TestExpressionParsing(t *testing.T) {
	tests := []struct{ expression, expected string }{
		{"1 + 2 * 3", "(1 + (2 * 3))"},
	}

	for _, tt := range tests {
		e := parseExpression(t, tt.expression)
		if s := e.String(); s != tt.expected {
			t.Errorf("expression '%s' not parsed correctly, expected %s but got %s",
				tt.expression, tt.expected, s)
		}
	}
}
