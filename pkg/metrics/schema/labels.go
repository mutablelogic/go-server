package schema

import (
	"fmt"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func LabelSet(labels ...any) ([]string, error) {
	labelset := make([]string, 0, len(labels)>>1)
	for i := 0; i < len(labels); i += 2 {
		key, ok := labels[i].(string)
		if !ok || key == "" {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid label: %q", labels[i])
		} else if !reLabelName.MatchString(key) {
			return nil, httpresponse.ErrBadRequest.Withf("Invalid label: %q", key)
		}
		labelset = append(labelset, fmt.Sprintf("%v=%q", key, labels[i+1]))
	}
	return labelset, nil
}
