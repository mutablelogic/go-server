package expr

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Scanner represents a lexical scanner.
type Scanner struct {
	r   *bufio.Reader
	pos Pos
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tokDefaultCapacity = 1024
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader, pos Pos) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(r),
		pos: pos,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Tokens returns the stream of tokens from the reader
//func (s *Scanner) Tokens()

// Scan returns the next token and literal value.
func (s *Scanner) Scan() *Token {
	// Read the next rune, and advance the position
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if unicode.IsSpace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if unicode.IsLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if unicode.IsDigit(ch) {
		s.unread()
		return s.scanDigit()
	} else if ch == '"' {
		s.unread()
		return s.scanString()
	} else if ch == '\'' {
		s.unread()
		return s.scanString()
	}

	// Otherwise read the individual character
	if kind, exists := tokenKindMap[ch]; exists {
		return NewToken(kind, string(ch), s.pos)
	} else {
		return nil
	}
}

// Tokens returns tokens from the scanner until EOF or illegal input is encountered
func (s *Scanner) Tokens() ([]*Token, error) {
	result := make([]*Token, 0, tokDefaultCapacity)
	for {
		tok := s.Scan()
		if tok == nil {
			return nil, NewPosError(ErrBadParameter.With("Illegal input"), s.pos)
		} else if tok.Kind == EOF {
			break
		}
		result = append(result, tok)
	}
	// Return tokens
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() *Token {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !unicode.IsSpace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return NewToken(Space, buf.String(), s.pos)
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() *Token {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Reserved words
	keyword := strings.ToUpper(buf.String())
	if kind, exists := tokenKeywordMap[keyword]; exists {
		return NewToken(kind, buf.String(), s.pos)
	}

	// Otherwise return as a regular identifier.
	return NewToken(Ident, buf.String(), s.pos)
}

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
func (s *Scanner) scanString() *Token {
	// Read the delimiter
	ending := s.read()

	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(ending)

	// Read every subsequent character into the buffer.
	for {
		if ch := s.read(); ch == eof {
			// Return nil if the string is not terminated
			return nil
		} else if ch == ending {
			break
		} else if ch == '\\' {
			// If the next character is an escape then write the escaped char
			next := s.read()
			if next == eof {
				// Unterminated escape
				return nil
			} else if next == 'n' {
				buf.WriteRune('\n')
			} else if next == 'r' {
				buf.WriteRune('\r')
			} else if next == '\\' {
				buf.WriteRune('\\')
			} else if next == '"' {
				buf.WriteRune('\\')
			} else if next == '\'' {
				buf.WriteRune('\'')
			} else {
				// Invalid escape
				return nil
			}
		} else {
			buf.WriteRune(ch)
		}
	}

	// Return the string
	return NewToken(String, buf.String(), s.pos)
}

// scanDigit consumes digits
func (s *Scanner) scanDigit() *Token {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent digit into the buffer.
	// Non-digit characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !unicode.IsDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier.
	return NewToken(Number, buf.String(), s.pos)
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	// Read a rune
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}

	// Mark previous position
	s.pos.x, s.pos.y = s.pos.Line, s.pos.Col

	// Advance position
	if ch == '\n' {
		s.pos.Line++
		s.pos.Col = 0
	} else if ch != eof {
		s.pos.Col++
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	// Restore previous position
	s.pos.Line, s.pos.Col, s.pos.x, s.pos.y = s.pos.x, s.pos.y, 0, 0
}
