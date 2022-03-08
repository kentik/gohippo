package hippo

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// test batch building when every populator is too big for a single batch
func TestBatchBuilder_FailureImpossible(t *testing.T) {
	a := require.New(t)

	maxSize := 10000
	sut := NewBatchBuilder(maxSize, true, 0)

	// 10 populators, 1000 IP addresses each - each populator is 11,630 bytes
	ips := buildIPAddresses(1000)
	for i := 0; i < 10; i++ {
		a.NoError(sut.AddUpsert(&TagUpsert{
			Value: fmt.Sprintf("abcdef_%d", i),
			Criteria: []TagCriteria{
				{
					Direction:   "asc",
					IPAddresses: ips,
				},
			},
		}))
	}

	batchBytes, err := sut.BuildBatch()
	a.Nil(batchBytes)
	a.Error(err)
	a.Equal("Have 10 remaining upserts, but could not fit any into the batch. The smallest one is 11630 bytes", err.Error())
}

// test building a batch from big and small upserts
// - also tests that batch guid is necessary on batch parts 2-N
func TestBatchBuilder_SuccessWithBigAndSmallUpserts(t *testing.T) {
	a := require.New(t)

	// picked a good size that spreads out big and little populators
	maxSize := 12220
	sut := NewBatchBuilder(maxSize, true, 13)

	runTest := func() {
		// 5 big populators 11,630 bytes each
		ips := buildIPAddresses(1000)
		for i := 0; i < 5; i++ {
			upsert := TagUpsert{
				Value: fmt.Sprintf("big_%d", i),
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: ips,
					},
				},
			}
			if i == 0 {
				// make sure the serialized upsert is the size we think
				batchBytes, err := json.Marshal(upsert)
				a.NoError(err)
				a.Equal(11627, len(batchBytes))
			}
			a.NoError(sut.AddUpsert(&upsert))
		}

		// 5 populators with 10 IP addresses each - 163 bytes per upsert
		ips = buildIPAddresses(10)
		for i := 0; i < 5; i++ {
			upsert := TagUpsert{
				Value: fmt.Sprintf("small_%d", i),
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: ips,
					},
				},
			}
			a.NoError(sut.AddUpsert(&upsert))

			if i == 0 {
				// make sure the serialized upsert is the size we think
				batchBytes, err := json.Marshal(upsert)
				a.NoError(err)
				a.Equal(162, len(batchBytes))
			}
		}

		verifyBatch := func(bigCount int, smallCount int, guid string, batchSizeBytes int, isComplete bool, ttlMinutes uint32) {
			batchBytes, err := sut.BuildBatch()
			a.NoError(err)
			a.NotNil(batchBytes)
			a.Equal(batchSizeBytes, len(batchBytes))
			a.True(len(batchBytes) <= maxSize)

			// make sure the batch is proper JSON
			batch := TagBatchPart{}
			a.NoError(json.Unmarshal(batchBytes, &batch))

			a.Equal(guid, batch.BatchGUID)
			a.Equal(isComplete, batch.IsComplete)
			a.Equal(ttlMinutes, batch.TTLMinutes)

			// make sure we have the right number of big and small upserts
			foundBigCount := 0
			foundSmallCount := 0
			for _, upsert := range batch.Upserts {
				a.Equal(1, len(upsert.Criteria))
				if strings.HasPrefix(upsert.Value, "small_") {
					foundSmallCount++
					a.Equal(10, len(upsert.Criteria[0].IPAddresses))
				} else if strings.HasPrefix(upsert.Value, "big_") {
					foundBigCount++
					a.Equal(1000, len(upsert.Criteria[0].IPAddresses))
				} else {
					a.FailNow("Invalid value: %s", upsert.Value)
				}
			}
			a.Equal(bigCount, foundBigCount)
			a.Equal(smallCount, foundSmallCount)
		}

		// 5 batches - each one has one big populator. First one has 3 small, second has 2 small
		verifyBatch(1, 3, "", 12193, false, 13)

		// now try to build a batch without providing GUID, and expect it'll fail
		batchBytes, err := sut.BuildBatch()
		a.Nil(batchBytes)
		a.Error(err)
		a.Equal("Only first batch may be sent without batch GUID", err.Error())

		// set the GUID and try again
		sut.SetBatchGUID("705e4dcb-3ecd-24f3-3a35-3e926e4bded5")

		verifyBatch(1, 2, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 12066, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11740, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11740, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11739, true, 13)

		// make sure that the batch builder knows it's done
		batchBytes, err = sut.BuildBatch()
		a.Nil(batchBytes)
		a.NoError(err)
	}

	// run the test twice to make sure we can reuse SUT
	runTest()
	sut.Reset(maxSize, true, 13)
	runTest()
}
