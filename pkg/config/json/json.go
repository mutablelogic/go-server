package json

import (
	"encoding/json"
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Unmarshal(data []byte, r any) error {
	return json.Unmarshal(data, r)
}
