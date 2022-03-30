package hippo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUint64RangeFromString(t *testing.T) {
	sut, errStr := NewFlexUint64RangesFromStrings([]string{"100", "33-36"})
	assert.Equal(t, "", errStr)
	assert.Equal(t, 2, len(sut))
	assert.Equal(t, uint64(100), sut[0].Start)
	assert.Equal(t, uint64(100), sut[0].End)
	assert.Equal(t, uint64(33), sut[1].Start)
	assert.Equal(t, uint64(36), sut[1].End)

	sut, errStr = NewFlexUint64RangesFromStrings([]string{" 100 "})
	assert.Equal(t, "", errStr)
	assert.Equal(t, 1, len(sut))
	assert.Equal(t, uint64(100), sut[0].Start)
	assert.Equal(t, uint64(100), sut[0].End)

	_, errStr = NewFlexUint64RangesFromStrings([]string{"a", "1"})
	assert.True(t, len(errStr) > 0)

	sut, errStr = NewFlexUint64RangesFromStrings([]string{"100-200", "3"})
	assert.Equal(t, "", errStr)
	assert.Equal(t, 2, len(sut))
	assert.Equal(t, uint64(100), sut[0].Start)
	assert.Equal(t, uint64(200), sut[0].End)
	assert.Equal(t, uint64(3), sut[1].Start)
	assert.Equal(t, uint64(3), sut[1].End)

	sut, errStr = NewFlexUint64RangesFromStrings([]string{"100 - 200"})
	assert.Equal(t, "", errStr)
	assert.Equal(t, 1, len(sut))
	assert.Equal(t, uint64(100), sut[0].Start)
	assert.Equal(t, uint64(200), sut[0].End)

	_, errStr = NewFlexUint64RangesFromStrings([]string{"200-100"})
	assert.True(t, len(errStr) > 0)
}
