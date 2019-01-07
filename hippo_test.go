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
