package ast

import (
	"fmt"
	"strings"

	"github.com/thacuber2a03/cymbal/lexer"
)

type (
	Node interface{}

	Declaration interface {
		Node
		declNode()
	}

	Statement interface {
		Node
		stmtNode()
		fmt.Stringer
	}

	Expression interface {
		Node
		exprNode()
		fmt.Stringer
	}
)

// Represents a single file.
type Program struct {
	Declarations []Declaration
}

// Represents the code in the reset vector,
// i.e. the code starting directly at 0x100 in the memory.
type MainDecl Block

func (_ *MainDecl) declNode() {}

// Represents a delimited block of code.
type Block struct {
	// The line that the start of the block (the left brace) is located in.
	// Start int

	Statements []Statement
}

func (_ *Block) stmtNode() {}
func (_ *Block) exprNode() {}

func (b *Block) String() string {
	var out strings.Builder

	stmts := []string{}
	for _, s:= range b.Statements {
		stmts = append(stmts, s.String())
	}

	out.WriteString("{ ")
	out.WriteString(strings.Join(stmts, "; "))
	out.WriteString(" }")
	return out.String()
}

// Represents a literal `<value> <port> DEO` sequence.
type DEOStatement struct {
	Port, Value Expression
}

func (_ *DEOStatement) stmtNode() {}

func (ds *DEOStatement) String() string {
	var out strings.Builder
	out.WriteString("deo ")
	out.WriteString(ds.Value.String())
	out.WriteString(", ")
	out.WriteString(ds.Port.String())
	return out.String()
}

// Represents a binary expression.
type Binary struct {
	Left, Right Expression
	Operator    lexer.Token
}

func (_ *Binary) exprNode() {}

func (b *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)",
		b.Left, b.Operator.Lexeme, b.Right)
}

// Represents a value.
type Literal struct {
	Value int16
}

func (_ *Literal) exprNode() {}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}
