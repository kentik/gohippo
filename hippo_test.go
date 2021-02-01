package hippo

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeTagBatchPart(t *testing.T) {

	c := NewHippo("agent", "email", "token")

	// Basic
	r := &TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts:    []TagUpsert{},
	}
	bAll, numUp, err := c.EncodeTagBatchPart(r)

	assert.NoError(t, err)
	assert.Equal(t, numUp, len(r.Upserts))
	for _, b := range bAll {
		assert.True(t, len(b) < MAX_HIPPO_SIZE, "Len: %d", len(b))
	}

	// Now, make sure that chunking is working
	devicename := make([]string, 100)
	for i := 0; i < 100; i++ {
		devicename[i] = "B"
	}
	deviceNameBase := strings.Join(devicename, "") + "_%d"
	r = &TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts:    make([]TagUpsert, MAX_HIPPO_SIZE/100),
	}

	for i := 0; i < MAX_HIPPO_SIZE/100; i++ {
		r.Upserts[i] = TagUpsert{
			Value: fmt.Sprintf("%d", i),
			Criteria: []TagCriteria{
				{
					Direction:         "dst",
					DeviceNameRegexes: []string{fmt.Sprintf(deviceNameBase, i)},
				},
			},
		}
	}

	bAllNew, numUp, err := c.EncodeTagBatchPart(r)
	assert.NoError(t, err)

	assert.Equal(t, len(r.Upserts), numUp, "Missing upserts")
	for _, b := range bAllNew {
		assert.True(t, len(b) < MAX_HIPPO_SIZE, "Len: %d; numparts: %d", len(b), len(bAllNew))
	}

	// One more with an odd number of segments expected
	r = &TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts:    make([]TagUpsert, MAX_HIPPO_SIZE/33),
	}

	for i := 0; i < MAX_HIPPO_SIZE/33; i++ {
		r.Upserts[i] = TagUpsert{
			Value: fmt.Sprintf("%d", i),
			Criteria: []TagCriteria{
				{
					Direction:         "dst",
					DeviceNameRegexes: []string{fmt.Sprintf(deviceNameBase, i)},
				},
			},
		}
	}

	bAllNew, numUp, err = c.EncodeTagBatchPart(r)
	assert.NoError(t, err)

	assert.Equal(t, len(r.Upserts), numUp, "Missing upserts")
	for _, b := range bAllNew {
		assert.True(t, len(b) < MAX_HIPPO_SIZE, "Len: %d; numparts: %d", len(b), len(bAllNew))
	}
}

// TestCompactTagBatchPart tests that compactTagBatchPart compacts one upsert with two upserts with the same value
// (but different cases) collapses them down into one with two rules
func TestCompactTagBatchPart(t *testing.T) {
	r := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			TagUpsert{
				Value: "my device",
				Criteria: []TagCriteria{
					{
						Direction:         "dst",
						DeviceNameRegexes: []string{"foo"},
					},
				},
			},
			TagUpsert{
				Value: "My device",
				Criteria: []TagCriteria{
					{
						Direction:         "dst",
						DeviceNameRegexes: []string{"bar"},
					},
				},
			},
		},
	}

	compactTagBatchPart := compactTagBatchPart(r)
	assert.True(t, compactTagBatchPart.ReplaceAll)
	assert.True(t, compactTagBatchPart.IsComplete)
	assert.Equal(t, 1, len(compactTagBatchPart.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	assert.Equal(t, "My device", compactTagBatchPart.Upserts[0].Value)
	assert.Equal(t, 2, len(compactTagBatchPart.Upserts[0].Criteria))
	assert.Equal(t, "dst", compactTagBatchPart.Upserts[0].Criteria[0].Direction)
	assert.Equal(t, "dst", compactTagBatchPart.Upserts[0].Criteria[1].Direction)
	assert.Equal(t, 1, len(compactTagBatchPart.Upserts[0].Criteria[0].DeviceNameRegexes))
	assert.Equal(t, 1, len(compactTagBatchPart.Upserts[0].Criteria[1].DeviceNameRegexes))
	assert.Equal(t, "foo", compactTagBatchPart.Upserts[0].Criteria[0].DeviceNameRegexes[0])
	assert.Equal(t, "bar", compactTagBatchPart.Upserts[0].Criteria[1].DeviceNameRegexes[0])
}

// TestCompactTagBatchPartNoCompact makes sure we don't combine two values that aren't the same
func TestCompactTagBatchPartNoCompact(t *testing.T) {
	r := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			TagUpsert{
				Value: "my device 1",
				Criteria: []TagCriteria{
					{
						Direction:         "dst",
						DeviceNameRegexes: []string{"foo"},
					},
				},
			},
			TagUpsert{
				Value: "My device 2",
				Criteria: []TagCriteria{
					{
						Direction:         "dst",
						DeviceNameRegexes: []string{"bar"},
					},
				},
			},
		},
	}

	compactTagBatchPart := compactTagBatchPart(r)
	assert.True(t, compactTagBatchPart.ReplaceAll)
	assert.True(t, compactTagBatchPart.IsComplete)
	assert.Equal(t, 2, len(compactTagBatchPart.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	assert.True(t, ("my device 1" == compactTagBatchPart.Upserts[0].Value && "My device 2" == compactTagBatchPart.Upserts[1].Value) ||
		("my device 1" == compactTagBatchPart.Upserts[1].Value && "My device 2" == compactTagBatchPart.Upserts[0].Value))
}

func TestFlexStringCriteriaEncoding(t *testing.T) {
	assert := assert.New(t)

	rule := TagCriteria{
		Direction: "either",
		Str00: []FlexStringCriteria{
			FlexStringCriteria{
				Action: FlexStringActionExact,
				Value:  "foo",
			},
			FlexStringCriteria{
				Action: FlexStringActionPrefix,
				Value:  "bar",
			},
		},
	}

	expect, err := json.MarshalIndent(map[string]interface{}{
		"direction": "either",
		"str00": []map[string]interface{}{
			map[string]interface{}{
				"action": "exact",
				"value":  "foo",
			},
			map[string]interface{}{
				"action": "prefix",
				"value":  "bar",
			},
		},
	}, "", "  ")
	assert.NoError(err)

	actual, err := json.MarshalIndent(&rule, "", "  ")
	assert.NoError(err)

	assert.Equal(string(expect), string(actual))
}
