package types_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Ptr_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("StringPtr", func(t *testing.T) {
		ptr1 := types.Ptr("")
		assert.NotNil(ptr1)
		assert.Equal(*ptr1, "")

		ptr2 := types.Ptr("hello")
		assert.NotNil(ptr2)
		assert.Equal(*ptr2, "hello")
	})

	t.Run("PtrString", func(t *testing.T) {
		str1 := types.Value[string](nil)
		assert.Equal(str1, "")

		str2 := "hello world"
		str3 := types.Value(&str2)
		assert.Equal(str2, str3)
	})

	t.Run("TrimStringPtr", func(t *testing.T) {
		ptr2 := types.TrimStringPtr(nil)
		assert.Nil(ptr2)

		str3 := "hello world"
		ptr4 := types.TrimStringPtr(&str3)
		assert.NotNil(ptr4)
		assert.Equal(str3, *ptr4)

		str5 := " hello world "
		ptr6 := types.TrimStringPtr(&str5)
		assert.NotNil(ptr6)
		assert.Equal("hello world", *ptr6)

		str7 := "   "
		ptr8 := types.TrimStringPtr(&str7)
		assert.Nil(ptr8)

		str9 := ""
		ptr10 := types.TrimStringPtr(&str9)
		assert.Nil(ptr10)
	})
}
