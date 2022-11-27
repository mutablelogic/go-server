package expr

import (
	"io"
	"strconv"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Parser represents an expression parser
type Parser struct {
	s   *Scanner
	t   []*Token
	cur int
	err error
}

type prefixParseFn func(*Parser) Node
type infixParseFn func(*Parser, Node) Node

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	prefixParseFns map[TokenKind]prefixParseFn
	infixParseFns  map[TokenKind]infixParseFn
)

func init() {
	prefixParseFns = map[TokenKind]prefixParseFn{
		Number: parseNumber,
		Plus:   parsePrefixPlus,
		Minus:  parsePrefixMinus,
		Not:    parsePrefixNot,
	}
	infixParseFns = map[TokenKind]infixParseFn{
		Plus: parseInfixPlus,
	}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewParser returns a new instance of Parser
func NewParser(r io.Reader, pos Pos) *Parser {
	this := &Parser{
		s: NewScanner(r, pos),
	}
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Parse is the main function call given a lexer instance it will parse
// and construct an abstract syntax tree for the given input.
func (p *Parser) Parse() (Node, error) {
	// Get tokens
	if tokens, err := p.s.Tokens(); err != nil {
		return nil, err
	} else if len(tokens) == 0 {
		return nil, NewPosError(ErrBadParameter.With("Empty input"), p.s.pos)
	} else {
		p.t = tokens
		p.cur = 0
		p.err = nil
	}

	// Cycle through tokens until we reach EOF
	result := &ExpressionListNode{}
	for kind := p.next(); kind != EOF; kind = p.next() {
		// Ignore any whitespace
		if kind == Space {
			continue
		}
		// Parse expression
		if node := p.parseExpression(); node == nil {
			break
		} else {
			result.v = append(result.v, node)
		}
	}

	// Return success
	return result, p.err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// returns the kind of the current token
func (p *Parser) next() TokenKind {
	if p.cur >= len(p.t) {
		return EOF
	}
	p.cur++
	return p.t[p.cur-1].Kind
}

// returns the kind of the next token, or EOF
func (p *Parser) peek() TokenKind {
	if p.cur >= len(p.t) {
		return EOF
	}
	return p.t[p.cur].Kind
}

// returns the current token kind
func (p *Parser) kind() TokenKind {
	return p.t[p.cur-1].Kind
}

// returns the current token value as a string
func (p *Parser) val() string {
	return p.t[p.cur-1].toString()
}

// returns the current token position
func (p *Parser) pos() Pos {
	return p.t[p.cur-1].Pos
}

// returns the current token position
func (p *Parser) token() *Token {
	return p.t[p.cur-1]
}

// parseExpression parses an prefix expression
func (p *Parser) parseExpression() Node {
	kind := p.kind()
	prefix, exists := prefixParseFns[kind]
	if !exists {
		p.err = multierror.Append(p.err, NewPosError(ErrBadParameter.Withf("Unexpected: %q", p.val()), p.pos()))
		return nil
	}

	// Evaluate left hand side
	left := prefix(p)
	if left == nil {
		return nil
	}

	// Wind beyond any spaces
	for kind := p.peek(); kind != EOF && kind != Space; kind = p.next() {
		// Continue consuming tokens
	}

	// Parse infix function
	if fn, exists := infixParseFns[kind]; exists {
		left = fn(p, left)
		if left == nil {
			return nil
		}
	}

	return left
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - PREFIX

// 0123455 => number
func parseNumber(p *Parser) Node {
	return parseNumberSign(p, true)
}

func parseNumberSign(p *Parser, plus bool) Node {
	val := toSign(plus) + p.val()

	// Check for <number><punkt> to interpret as a float
	if p.peek() == Punkt {
		return parseFloat(p, val)
	} else if plus {
		if n, err := strconv.ParseUint(val, 0, 64); err != nil {
			p.err = multierror.Append(p.err, NewParseError(p.token()))
			return nil
		} else {
			return &UintNumberNode{n}
		}
	} else {
		if n, err := strconv.ParseInt(val, 0, 64); err != nil {
			p.err = multierror.Append(p.err, NewParseError(p.token()))
			return nil
		} else {
			return &IntNumberNode{n}
		}
	}
}

// 0123.455 => float number
// 0123. => float number
func parseFloat(p *Parser, val string) Node {
	if kind := p.next(); kind != Punkt {
		p.err = multierror.Append(p.err, NewParseError(p.token()))
		return nil
	}
	if p.peek() == Number {
		p.next()
		val += "." + p.val()
	}
	if n, err := strconv.ParseFloat(val, 64); err != nil {
		p.err = multierror.Append(p.err, NewPosError(err, p.pos()))
		return nil
	} else {
		return &FloatNumberNode{n}
	}
}

// +0123455 => number
func parsePrefixPlus(p *Parser) Node {
	// Check for <number>. to interpret as a number
	if p.peek() == Number {
		p.next()
		return parseNumberSign(p, true)
	} else {
		p.err = multierror.Append(p.err, NewParseError(p.token()))
		return nil
	}
}

// -0123455 => number
func parsePrefixMinus(p *Parser) Node {
	// Check for <number>. to interpret as a number
	if p.peek() == Number {
		p.next()
		return parseNumberSign(p, false)
	} else {
		p.err = multierror.Append(p.err, NewParseError(p.token()))
		return nil
	}
}

// !<Expression> => BitwiseNot(expression)
func parsePrefixNot(p *Parser) Node {
	p.next()
	if node := p.parseExpression(); node == nil {
		return nil
	} else {
		return &UnaryFunction{Not, node}
	}
}

// <Expression> + <Expression>
func parseInfixPlus(p *Parser, left Node) Node {
	p.err = multierror.Append(p.err, NewPosError(ErrNotImplemented, p.pos()))
	return nil
}

func toSign(plus bool) string {
	if plus {
		return ""
	}
	return "-"
}
