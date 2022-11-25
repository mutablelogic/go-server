package types

import (
	"bufio"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Evaluate an expression into a value, anything within a ${} is evaluated
// as an expression, the variables and functions are embedded within the
// context.
/*func (v *Eval) Eval(parent context.Context, path string) (any, error) {
	if tokens, err := v.Tokenize(); err != nil {
		return nil, err
	} else {
		fmt.Printf("PARSE %q => %q\n", string(*v), tokens)
	}
	return string(*v), nil
}*/

// Interpolation of a string to return a set of tokens, some of which are
// strings, and some of which are expression nodes
func (v *Eval) Interpolate() ([]*Node, error) {
	tokens, err := v.InterpolateTokenize()
	if err != nil {
		return nil, err
	}
	// Now parse the strings, marking out specifically ${ $${ and } characters
	result := make([]*Node, 0, len(tokens))
	state := 0
	for _, token := range tokens {
		switch state {
		case 0:
			if token == "${" {
				state = 1
			} else if token == "$${" {
				result = append(result, NewString("${"))
			} else {
				result = append(result, NewString(token))
			}
		case 1:
			if token == "}" {
				state = 0
			} else {
				result = append(result, NewExpr(token))
			}
		}
	}

	// If we aren't in state 0 when finished, then we are missing a final }
	if state != 0 {
		return nil, ErrUnexpectedResponse.Withf("Missing }")
	}

	// Return list of nodes otherwise
	return result, nil
}

// Interpolation of a string to return a set of string tokens, which
// are either "${" "$${" or "}" to be recognized as special, or
// any other string which is interpreted based on context as string
// or expression
func (v *Eval) InterpolateTokenize() ([]string, error) {
	tokens := make([]string, 0, len(*v))
	r := strings.NewReader(string(*v))
	scanner := bufio.NewScanner(r)
	scanner.Split(scanInterpolateTokens)
	for {
		if scanned := scanner.Scan(); !scanned {
			break
		}
		if token := scanner.Text(); token != "" {
			tokens = append(tokens, token)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func scanInterpolateTokens(data []byte, atEOF bool) (int, []byte, error) {
	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Check for $${ and return if found
	if i := strings.Index(string(data), "$${"); i == 0 {
		return 3, data[0:3], nil
	} else if i > 0 {
		return i, data[0:i], nil
	}
	// Check for ${ and return if found
	if i := strings.Index(string(data), "${"); i == 0 {
		return 2, data[0:2], nil
	} else if i > 0 {
		return i, data[0:i], nil
	}
	// Check for } and return if found
	if i := strings.Index(string(data), "}"); i == 0 {
		return 1, data[0:1], nil
	} else if i > 0 {
		return i, data[0:i], nil
	}
	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}
	// Request more data
	return 0, nil, nil
}
