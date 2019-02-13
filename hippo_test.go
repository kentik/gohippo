package hippo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeReq(t *testing.T) {

	c := NewHippo("agent", "email", "token")

	// Basic
	r := &Req{
		Replace:  true,
		Complete: true,
		Upserts:  []Upsert{},
	}
	bAll, numUp, err := c.EncodeReq(r)

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
	r = &Req{
		Replace:  true,
		Complete: true,
		Upserts:  make([]Upsert, MAX_HIPPO_SIZE/100),
	}

	for i := 0; i < MAX_HIPPO_SIZE/100; i++ {
		r.Upserts[i] = Upsert{
			Val: fmt.Sprintf("%d", i),
			Rules: []Rule{
				{
					Dir:         "dst",
					DeviceNames: []string{fmt.Sprintf(deviceNameBase, i)},
				},
			},
		}
	}

	bAllNew, numUp, err := c.EncodeReq(r)
	assert.NoError(t, err)

	assert.Equal(t, len(r.Upserts), numUp, "Missing upserts")
	for _, b := range bAllNew {
		assert.True(t, len(b) < MAX_HIPPO_SIZE, "Len: %d; numparts: %d", len(b), len(bAllNew))
	}

	// One more with an odd number of segments expected
	r = &Req{
		Replace:  true,
		Complete: true,
		Upserts:  make([]Upsert, MAX_HIPPO_SIZE/33),
	}

	for i := 0; i < MAX_HIPPO_SIZE/33; i++ {
		r.Upserts[i] = Upsert{
			Val: fmt.Sprintf("%d", i),
			Rules: []Rule{
				{
					Dir:         "dst",
					DeviceNames: []string{fmt.Sprintf(deviceNameBase, i)},
				},
			},
		}
	}

	bAllNew, numUp, err = c.EncodeReq(r)
	assert.NoError(t, err)

	assert.Equal(t, len(r.Upserts), numUp, "Missing upserts")
	for _, b := range bAllNew {
		assert.True(t, len(b) < MAX_HIPPO_SIZE, "Len: %d; numparts: %d", len(b), len(bAllNew))
	}
}

// TestCompactReq tests that compactReq compacts one upsert with two upserts with the same value
// (but different cases) collapses them down into one with two rules
func TestCompactReq(t *testing.T) {
	r := Req{
		Replace:  true,
		Complete: true,
		Upserts: []Upsert{
			Upsert{
				Val: "my device",
				Rules: []Rule{
					{
						Dir:         "dst",
						DeviceNames: []string{"foo"},
					},
				},
			},
			Upsert{
				Val: "My device",
				Rules: []Rule{
					{
						Dir:         "dst",
						DeviceNames: []string{"bar"},
					},
				},
			},
		},
	}

	compactReq := compactReq(r)
	assert.True(t, compactReq.Replace)
	assert.True(t, compactReq.Complete)
	assert.Equal(t, 1, len(compactReq.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	assert.Equal(t, "My device", compactReq.Upserts[0].Val)
	assert.Equal(t, 2, len(compactReq.Upserts[0].Rules))
	assert.Equal(t, "dst", compactReq.Upserts[0].Rules[0].Dir)
	assert.Equal(t, "dst", compactReq.Upserts[0].Rules[1].Dir)
	assert.Equal(t, 1, len(compactReq.Upserts[0].Rules[0].DeviceNames))
	assert.Equal(t, 1, len(compactReq.Upserts[0].Rules[1].DeviceNames))
	assert.Equal(t, "foo", compactReq.Upserts[0].Rules[0].DeviceNames[0])
	assert.Equal(t, "bar", compactReq.Upserts[0].Rules[1].DeviceNames[0])
}

// TestCompactReqNoCompact makes sure we don't combine two values that aren't the same
func TestCompactReqNoCompact(t *testing.T) {
	r := Req{
		Replace:  true,
		Complete: true,
		Upserts: []Upsert{
			Upsert{
				Val: "my device 1",
				Rules: []Rule{
					{
						Dir:         "dst",
						DeviceNames: []string{"foo"},
					},
				},
			},
			Upsert{
				Val: "My device 2",
				Rules: []Rule{
					{
						Dir:         "dst",
						DeviceNames: []string{"bar"},
					},
				},
			},
		},
	}

	compactReq := compactReq(r)
	assert.True(t, compactReq.Replace)
	assert.True(t, compactReq.Complete)
	assert.Equal(t, 2, len(compactReq.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	assert.True(t, ("my device 1" == compactReq.Upserts[0].Val && "My device 2" == compactReq.Upserts[1].Val) ||
		("my device 1" == compactReq.Upserts[1].Val && "My device 2" == compactReq.Upserts[0].Val))
}
