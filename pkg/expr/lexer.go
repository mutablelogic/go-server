package expr

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reSpace  = regexp.MustCompile(`\s+`)
	reIdent  = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_-]*`)
	reNumber = regexp.MustCompile(`([0-9]+)(\.[0-9]+)?`)
	rePunkt  = regexp.MustCompile(`\.`)
)

var (
	reMap = map[*regexp.Regexp]TokenKind{
		reSpace:  Space,
		reIdent:  Ident,
		reNumber: Number,
		rePunkt:  Punkt,
	}
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Expression Lexer converts from stream of bytes into a set of lexical tokens
// which are one of ident, number, string, delimiter or operator
func Lexer(r io.Reader, pos Pos) ([]*Token, error) {
	// Split on token boundaries
	tokens, err := lexerSplit(r, capTokens)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Tokens: %q\n", tokens)

	return nil, nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Interpolation of a string to return a set of strings, which
// are either "${" "$${", "}", newline to be recognized as special, or
// any other string which is interpreted based on context as string
// or expression
func lexerSplit(r io.Reader, cap int) ([]string, error) {
	tokens := make([]string, 0, cap)
	scanner := bufio.NewScanner(r)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		// Return nothing if at end of file and no data passed
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// Find the first token
		for re, kind := range reMap {
			if i := re.FindSubmatchIndex(data); i == nil {
				continue
			} else if i[0] == 0 {
				fmt.Printf("  %q Found %v %v\n", string(data), kind, i)
				return i[1], data[i[0]:i[1]], nil
			}
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
