package types

import (
	"strconv"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Task) UnmarshalJSON(data []byte) error {
	if v, err := strconv.Unquote(string(data)); err != nil {
		return err
	} else {
		t.Ref = v
		return nil
	}
}
