package types_test

import (
	"encoding/json"
	"testing"

	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_task_000(t *testing.T) {
	var task types.Task

	assert := assert.New(t)
	err := json.Unmarshal([]byte(`"ref"`), &task)
	assert.NoError(err)
	assert.Equal("ref", task.Ref)
}
