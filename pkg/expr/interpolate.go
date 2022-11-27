package expr

import (
	"bufio"
	"io"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type tokenArray []*Token

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tokQuoteExpr = "$${" // $${ => ${
	tokStartExpr = "${"  // ${ => start of expression
	tokEndExpr   = "}"   // $ => end of expression
	tokNewline   = "\n"  // \n => end of line
	capTokens    = 100   // Default capacity for tokens
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Interpolation of a stream to return a set of tokens, some of which are
// strings, and some of which are expression tokens
func Interpolate(r io.Reader, pos Pos) ([]*Token, error) {
	// Split on token boundaries
	tokens, err := interpolateSplit(r, capTokens)
	if err != nil {
		return nil, err
	}

	// Return nil if no tokens
	if len(tokens) == 0 {
		return nil, nil
	}

	// True if we are "in" an expression
	var inExpr bool
	var result = make(tokenArray, 0, len(tokens))

	// Now parse the strings, marking out specifically ${ $${ and } characters
	for _, token := range tokens {
		// Advance position
		if token == tokNewline {
			pos.Col = 0
			pos.Line++
		} else {
			pos.Col += uint(len(token))
		}
		// Eject tokens
		switch inExpr {
		case false:
			if token == tokStartExpr {
				inExpr = true
			} else if token == tokQuoteExpr {
				result = result.append(String, tokStartExpr, pos)
			} else {
				result = result.append(String, token, pos)
			}
		case true:
			if token == tokEndExpr {
				inExpr = false
			} else {
				result = result.append(Expr, token, pos)
			}
		}
	}

	// If we aren't in state 0 when finished, then we are missing a final }
	if inExpr {
		return nil, ErrUnexpectedResponse.Withf("Missing }")
	}

	// Return list of nodes otherwise
	return result, nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Interpolation of a string to return a set of strings, which
// are either "${" "$${", "}", newline to be recognized as special, or
// any other string which is interpreted based on context as string
// or expression
func interpolateSplit(r io.Reader, cap int) ([]string, error) {
	tokens := make([]string, 0, cap)
	scanner := bufio.NewScanner(r)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		// Return nothing if at end of file and no data passed
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// Find the first token
		var j int
		for _, tok := range []string{tokQuoteExpr, tokStartExpr, tokEndExpr, tokNewline} {
			// Check for tok and return if found
			if i := strings.Index(string(data), tok); i == 0 {
				return len(tok), data[0:len(tok)], nil
			} else if i > 0 {
				if j == 0 || j > i {
					j = i
				}
			}
		}
		// Return data up until the furst token
		if j > 0 {
			return j, data[0:j], nil
		}
		// If at end of file with data return the data
		if atEOF {
			return len(data), data, nil
		}
		// Request more data
		return 0, nil, nil
	})
	for {
		if scanned := scanner.Scan(); !scanned {
			break
		}
		if token := scanner.Text(); token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens, scanner.Err()
}

func (t tokenArray) append(k TokenKind, str string, pos Pos) tokenArray {
	if len(t) == 0 {
		t = append(t, NewToken(k, str, pos))
	} else if last := t[len(t)-1]; last.Kind == String {
		last.Val = last.Val.(string) + str
	} else {
		t = append(t, NewToken(k, str, pos))
	}
	return t
}
