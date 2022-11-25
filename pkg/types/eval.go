package types

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	// Package imports
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Evaluate an expression into a value, anything within a ${} is evaluated
// as an expression, the variables and functions are embedded within the
// context.
func (v *Eval) Eval(parent context.Context, path string) (any, error) {
	if tokens, err := v.Tokenize(); err != nil {
		return nil, err
	} else {
		fmt.Printf("PARSE %q => %q\n", string(*v), tokens)
	}
	return string(*v), nil
}

// Evaluate an expression into a value, anything within a ${} is evaluated
// as an expression, the variables and functions are embedded within the
// context.
func (v *Eval) Tokenize() ([]string, error) {
	tokens := make([]string, 0, len(*v))
	r := strings.NewReader(string(*v))
	scanner := bufio.NewScanner(r)
	scanner.Split(scanTokens)
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

func scanTokens(data []byte, atEOF bool) (int, []byte, error) {
	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
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
