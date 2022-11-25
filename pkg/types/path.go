package types

import (
	"os"
	"strconv"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Path) UnmarshalJSON(data []byte) error {
	if v, err := strconv.Unquote(string(data)); err != nil {
		return err
	} else {
		*p = Path(v)
		return nil
	}
}

func (p *Path) AbsFile() (string, error) {
	if stat, err := os.Stat(string(*p)); err != nil {
		return "", err
	} else if !stat.Mode().IsRegular() {
		return "", ErrBadParameter.Withf("config is not a regular file: %q", string(*p))
	} else {
		return string(*p), nil
	}
}

func (p *Path) AbsDir() (string, error) {
	if stat, err := os.Stat(string(*p)); err != nil {
		return "", err
	} else if !stat.Mode().IsDir() {
		return "", ErrBadParameter.Withf("config is not a directory: %q", string(*p))
	} else {
		return string(*p), nil
	}
}
