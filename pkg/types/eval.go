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
	fmt.Printf("PARSE %q\n", string(*v))
	r := strings.NewReader(string(*v))
	scanner := bufio.NewScanner(r)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Return nothing if at end of file and no data passed
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "${"); i >= 0 {
			return i + 2, data[0:i], nil
		}
		if i := strings.Index(string(data), "}"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		// If at end of file with data return the data
		if atEOF {
			return len(data), data, nil
		}
		return
	})
	for {
		scanned := scanner.Scan()
		if !scanned {
			break
		}
		line := scanner.Text()
		fmt.Printf(" TOK %q\n", line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return string(*v), nil
}
