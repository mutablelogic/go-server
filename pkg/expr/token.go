package expr

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Classifies the kind of token
type TokenKind uint

// Token is decomposed from []byte stream to represent a kind of
// token and the vaoue of the token
type Token struct {
	Kind TokenKind
	Val  any
	Pos  Pos
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Any TokenKind = iota
	String
	Expr
	Space
	Ident
	Number
	Punkt
	Question
	Colon
	Comma
	OpenParen
	CloseParen
	Equal
	Not
	Plus
	Minus
	Multiply
	Divide
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewToken(kind TokenKind, val any, pos Pos) *Token {
	return &Token{Kind: kind, Val: val, Pos: pos}
}

func NewStringToken(val string, pos Pos) *Token {
	return &Token{Kind: String, Val: val, Pos: pos}
}

func NewExprToken(val string, pos Pos) *Token {
	return &Token{Kind: Expr, Val: val, Pos: pos}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (k TokenKind) String() string {
	switch k {
	case Any:
		return "Any"
	case String:
		return "String"
	case Expr:
		return "Expr"
	case Space:
		return "Space"
	case Ident:
		return "Ident"
	case Number:
		return "Number"
	case Punkt:
		return "Punkt"
	default:
		return "[?? Invalid TokenKind value]"
	}
}

func (t *Token) String() string {
	switch t.Kind {
	case String:
		return fmt.Sprintf("str<%q>", t.Val)
	case Expr:
		return fmt.Sprintf("expr<%q>", t.Val)
	case Space:
		return "space<>"
	case Ident:
		return fmt.Sprintf("ident<%q>", t.Val)
	case Number:
		return fmt.Sprintf("number<%q>", t.Val)
	default:
		return fmt.Sprintf("any<%v>", t.Val)
	}
}
