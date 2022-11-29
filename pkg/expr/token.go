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
	OpenSquare
	CloseSquare
	OpenBrace
	CloseBrace
	Ampersand
	Equal
	Less
	Greater
	Plus
	Minus
	Multiply
	Divide
	Not
	True
	False
	Null
	EOF
	// The following types are used for nodes only
	List
	Lowest = Equal // Lowest precedence
)

var (
	eof = rune(0)
	// Special characters
	tokenKindMap = map[rune]TokenKind{
		'.': Punkt,
		'?': Question,
		':': Colon,
		',': Comma,
		'(': OpenParen,
		')': CloseParen,
		'[': OpenSquare,
		']': CloseSquare,
		'{': OpenBrace,
		'}': CloseBrace,
		'&': Ampersand,
		'=': Equal,
		'<': Less,
		'>': Greater,
		'!': Not,
		'+': Plus,
		'-': Minus,
		'*': Multiply,
		'/': Divide,
		eof: EOF,
	}
	// Reserved words
	tokenKeywordMap = map[string]TokenKind{
		"true":  True,
		"false": False,
		"null":  Null,
	}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewToken(kind TokenKind, val any, pos Pos) *Token {
	return &Token{Kind: kind, Val: val, Pos: pos}
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
	case Question:
		return "Question"
	case Colon:
		return "Colon"
	case Comma:
		return "Comma"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case OpenSquare:
		return "OpenSquare"
	case CloseSquare:
		return "CloseSquare"
	case OpenBrace:
		return "OpenBrace"
	case CloseBrace:
		return "CloseBrace"
	case Ampersand:
		return "Ampersand"
	case Equal:
		return "Equal"
	case Less:
		return "Less"
	case Greater:
		return "Greater"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Multiply:
		return "Multiply"
	case Divide:
		return "Divide"
	case Not:
		return "Not"
	case True:
		return "True"
	case False:
		return "False"
	case Null:
		return "Null"
	case EOF:
		return "EOF"
	case List:
		return "List"
	default:
		return "[?? Invalid TokenKind value]"
	}
}

func (t *Token) String() string {
	if t.Val == nil {
		return fmt.Sprintf("%v<nil>", t.Kind)
	}
	switch v := t.Val.(type) {
	case string:
		return fmt.Sprintf("%v<%q>", t.Kind, v)
	default:
		return fmt.Sprintf("%v<%v>", t.Kind, v)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Token) toString() string {
	if t.Val == nil {
		return "<nil>"
	}
	switch v := t.Val.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}
