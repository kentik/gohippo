package hippo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseASNs(t *testing.T) {
	var errMsg string
	ranges := ParseASNs([]string{"1 ", "2", "3-5 ", " 5-5", "6 - 5 "}, &errMsg)
	assert.Equal(t, 4, len(ranges))
	assert.True(t, len(errMsg) > 0)

	assert.Equal(t, uint32(1), ranges[0].Start)
	assert.Equal(t, uint32(1), ranges[0].End)
	assert.Equal(t, uint32(2), ranges[1].Start)
	assert.Equal(t, uint32(2), ranges[1].End)

	assert.Equal(t, uint32(3), ranges[2].Start)
	assert.Equal(t, uint32(5), ranges[2].End)

	assert.Equal(t, uint32(5), ranges[3].Start)
	assert.Equal(t, uint32(5), ranges[3].End)
}

func TestParsePorts(t *testing.T) {
	var errMsg string
	ranges := ParsePorts([]string{"1", "2", "3-5", "5-5", "6 - 5 "}, &errMsg)
	assert.Equal(t, 4, len(ranges))
	assert.True(t, len(errMsg) > 0)

	assert.Equal(t, uint32(1), ranges[0].Start)
	assert.Equal(t, uint32(1), ranges[0].End)
	assert.Equal(t, uint32(2), ranges[1].Start)
	assert.Equal(t, uint32(2), ranges[1].End)

	assert.Equal(t, uint32(3), ranges[2].Start)
	assert.Equal(t, uint32(5), ranges[2].End)

	assert.Equal(t, uint32(5), ranges[3].Start)
	assert.Equal(t, uint32(5), ranges[3].End)
}

func TestParseVLans(t *testing.T) {
	var errMsg string
	ranges := ParseVLans([]string{"-1", "4094-4096", "4094-4095", "1", "4095", "4096", "-1-5", "100-200", "37", "100-130", "5-4"}, &errMsg)
	assert.Equal(t, 6, len(ranges))
	assert.True(t, errMsg != "")

	assert.Equal(t, uint32(4094), ranges[0].Start)
	assert.Equal(t, uint32(4095), ranges[0].End)

	assert.Equal(t, uint32(1), ranges[1].Start)
	assert.Equal(t, uint32(1), ranges[1].End)

	assert.Equal(t, uint32(4095), ranges[2].Start)
	assert.Equal(t, uint32(4095), ranges[2].End)

	assert.Equal(t, uint32(100), ranges[3].Start)
	assert.Equal(t, uint32(200), ranges[3].End)

	assert.Equal(t, uint32(37), ranges[4].Start)
	assert.Equal(t, uint32(37), ranges[4].End)

	assert.Equal(t, uint32(100), ranges[5].Start)
	assert.Equal(t, uint32(130), ranges[5].End)

	// now sort and reevaluate
	VLanRangesSlice(ranges).Sort()
	assert.Equal(t, 6, len(ranges))

	assert.Equal(t, uint32(1), ranges[0].Start)
	assert.Equal(t, uint32(1), ranges[0].End)

	assert.Equal(t, uint32(37), ranges[1].Start)
	assert.Equal(t, uint32(37), ranges[1].End)

	assert.Equal(t, uint32(100), ranges[2].Start)
	assert.Equal(t, uint32(130), ranges[2].End)

	assert.Equal(t, uint32(100), ranges[3].Start)
	assert.Equal(t, uint32(200), ranges[3].End)

	assert.Equal(t, uint32(4094), ranges[4].Start)
	assert.Equal(t, uint32(4095), ranges[4].End)

	assert.Equal(t, uint32(4095), ranges[5].Start)
	assert.Equal(t, uint32(4095), ranges[5].End)

	assert.Equal(t, "1,37,100-130,100-200,4094-4095,4095", VLanRangesSlice(ranges).String())

	stringArray := VLanRangesSlice(ranges).ToStringArray()
	assert.Equal(t, 6, len(stringArray))
	assert.Equal(t, "1", stringArray[0])
	assert.Equal(t, "37", stringArray[1])
	assert.Equal(t, "100-130", stringArray[2])
	assert.Equal(t, "100-200", stringArray[3])
	assert.Equal(t, "4094-4095", stringArray[4])
	assert.Equal(t, "4095", stringArray[5])
}
