package hippo

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// test batch building when every populator is too big for a single batch
// - in this case, we send batches anyway, from the smallest to the biggest,
//   just in case server can accept them, we wanna send out whatever we can
func TestBatchBuilder_FailureOverLimit(t *testing.T) {
	a := require.New(t)

	maxSize := 10000
	sut := NewBatchBuilder(maxSize, true, 0)

	stringOfLength := func(length int) string {
		ret := ""
		for i := 0; i < length; i++ {
			ret += "a"
		}
		return ret
	}

	// 3 populators, 1000 IP addresses each - each populator is 11,629-11,631 bytes
	ips := buildIPAddresses(1000)
	for i := 0; i < 3; i++ {
		a.NoError(sut.AddUpsert(&TagUpsert{
			Value: fmt.Sprintf("big_%d_%s", i, stringOfLength(i+1)),
			Criteria: []TagCriteria{
				{
					Direction:   "src",
					IPAddresses: ips,
				},
			},
		}))
	}

	// 3 tiny populators that have no problem fitting into the first batch
	for i := 0; i < 3; i++ {
		a.NoError(sut.AddUpsert(&TagUpsert{
			Value: fmt.Sprintf("small_%d_%s", i, stringOfLength(i+1)),
			Criteria: []TagCriteria{
				{
					Direction:   "src",
					IPAddresses: []string{"5.1.5.1"},
				},
			},
		}))
	}

	verify := func(batchGUID string, hasUpsert bool, upsertValue string) {
		batchBytes, receivedUpsertCount, err := sut.BuildBatchRequest()
		a.NoError(err)
		if !hasUpsert {
			a.Nil(batchBytes)
			a.Equal(0, receivedUpsertCount)
			return
		}

		a.NotNil(batchBytes)
		a.True(len(batchBytes) > maxSize) // we had no choice but to build a bigger-than-desired batch
		a.Equal(1, receivedUpsertCount)
		batch := TagBatchPart{}
		a.NoError(json.Unmarshal(batchBytes, &batch))
		a.Equal(batchGUID, batch.BatchGUID)
		a.Equal(1, len(batch.Upserts))
		a.Equal(1, len(batch.Upserts[0].Criteria))
		a.Equal(1000, len(batch.Upserts[0].Criteria[0].IPAddresses))
		a.Equal(upsertValue, batch.Upserts[0].Value)
	}

	// 1/4 parts - contains all the small populators
	batchBytes, upsertCount, err := sut.BuildBatchRequest()
	a.NotNil(batchBytes)
	a.NoError(err)
	a.Equal(3, upsertCount)
	batch := TagBatchPart{}
	a.NoError(json.Unmarshal(batchBytes, &batch))
	a.Equal(3, len(batch.Upserts)) // 3 small ones

	// these should be ordered from smallest to biggest, because of the 3 bigger batches
	a.Equal("small_0_a", batch.Upserts[0].Value)
	a.Equal("small_1_aa", batch.Upserts[1].Value)
	a.Equal("small_2_aaa", batch.Upserts[2].Value)

	// make sure that we get an error when building the second part without a GUID
	batchBytes, upsertCount, err = sut.BuildBatchRequest()
	a.Nil(batchBytes)
	a.Zero(upsertCount)
	a.Error(err)
	a.Equal("Only first batch may be sent without batch GUID", err.Error())

	// set the GUID and continue
	guid := "805e4dcb-3ecd-24f3-3a35-3e926e4bded5"
	sut.SetBatchGUID(guid)

	// rest of the batches hold a single, oversized populator each, hoping for the best on the server side
	// should be ordered smallest to biggest, to get as much into the server as possible
	verify(guid, true, "big_0_a")
	verify(guid, true, "big_1_aa")
	verify(guid, true, "big_2_aaa")
	verify("", false, "")
}

// test building an empty replace-all bach does build an empty batch - once
func TestBatchBuilder_SuccessEmptyReplaceAllBatch(t *testing.T) {
	a := require.New(t)

	sut := NewBatchBuilder(3000000, true, 17)
	sut.SetSenderInfo(TagBatchPartSender{
		ServiceName:     "service-name",
		ServiceInstance: "service-instance",
		HostName:        "host-name",
	})

	// build the batch - make sure EVERYTHING is as we expect
	batchBytes, upsertCount, err := sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(0, upsertCount)
	a.NotNil(batchBytes)
	a.Equal(179, len(batchBytes))

	expectedBatch := TagBatchPart{
		BatchGUID:  "",
		ReplaceAll: true,
		IsComplete: true,
		Upserts:    []TagUpsert{},
		TTLMinutes: 17,
		Sender: TagBatchPartSender{
			ServiceName:     "service-name",
			ServiceInstance: "service-instance",
			HostName:        "host-name",
		},
	}

	receivedBatch := TagBatchPart{}
	a.NoError(json.Unmarshal(batchBytes, &receivedBatch))
	a.True(expectedBatch.Equal(receivedBatch))

	// try sending it again - should result in no batch, since the replace_all=true batch was just built
	batchBytes, upsertCount, err = sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(0, upsertCount)
	a.Nil(batchBytes)
}

// test building an empty non-replace-all bach doesn't build a batch
func TestBatchBuilder_SuccessEmptyNonReplaceAllBatch(t *testing.T) {
	a := require.New(t)

	sut := NewBatchBuilder(3000000, false, 17)
	sut.SetSenderInfo(TagBatchPartSender{
		ServiceName:     "service-name",
		ServiceInstance: "service-instance",
		HostName:        "host-name",
	})

	// attempt to build the batch - should do nothing, since there's no upserts, and this isn't a replace-all batch
	batchBytes, upsertCount, err := sut.BuildBatchRequest()
	a.NoError(err)
	a.Nil(batchBytes)
	a.Equal(0, upsertCount)

	// try sending it again - should result in no batch again
	batchBytes, upsertCount, err = sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(0, upsertCount)
	a.Nil(batchBytes)
}

// Test building a batch that fits in one part
func TestBatchBuilder_SuccessOneBatch(t *testing.T) {
	a := require.New(t)

	sut := NewBatchBuilder(3000000, true, 17)
	sut.SetSenderInfo(TagBatchPartSender{
		ServiceName:     "service-name",
		ServiceInstance: "service-instance",
		HostName:        "host-name",
	})

	// build a name for an upsert by index where the earlier indexes names are longer, so we can predict the order in the output
	nameFor := func(index int) string {
		ret := "value"
		for i := 1; i <= index; i++ {
			ret += "_A"
		}
		return ret + fmt.Sprintf("_%d", index)
	}

	for i := 1; i <= 5; i++ {
		upsert := TagUpsert{
			Value: nameFor(i),
			Criteria: []TagCriteria{
				{
					Direction:   "src",
					IPAddresses: []string{fmt.Sprintf("1.2.3.%d", i)},
				},
			},
		}
		a.NoError(sut.AddUpsert(&upsert))
	}

	// build the batch - make sure EVERYTHING is as we expect
	batchBytes, upsertCount, err := sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(5, upsertCount)
	a.Equal(568, len(batchBytes))

	expectedBatch := TagBatchPart{
		BatchGUID:  "",
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			TagUpsert{
				Value: "value_A_A_A_A_A_5",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.5"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_A_A_A_4",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_A_A_3",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.3"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_A_2",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.2"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_1",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.1"},
					},
				},
			},
		},
		TTLMinutes: 17,
		Sender: TagBatchPartSender{
			ServiceName:     "service-name",
			ServiceInstance: "service-instance",
			HostName:        "host-name",
		},
	}

	receivedBatch := TagBatchPart{}
	a.NoError(json.Unmarshal(batchBytes, &receivedBatch))

	a.True(expectedBatch.Equal(receivedBatch))
}

// Test building a batch that fits in 2 parts, with sender info
func TestBatchBuilder_SuccessTwoBatches(t *testing.T) {
	a := require.New(t)

	// single batch fits in 568 bytes - force 2 batches by setting max size to 500
	sut := NewBatchBuilder(500, true, 17)
	sut.SetSenderInfo(TagBatchPartSender{
		ServiceName:     "service-name",
		ServiceInstance: "service-instance",
		HostName:        "host-name",
	})

	// build a name for an upsert by index where the earlier indexes names are longer, so we can predict the order in the output
	nameFor := func(index int) string {
		ret := "value"
		for i := 1; i <= index; i++ {
			ret += "_A"
		}
		return ret + fmt.Sprintf("_%d", index)
	}

	for i := 1; i <= 5; i++ {
		upsert := TagUpsert{
			Value: nameFor(i),
			Criteria: []TagCriteria{
				{
					Direction:   "src",
					IPAddresses: []string{fmt.Sprintf("1.2.3.%d", i)},
				},
			},
		}
		a.NoError(sut.AddUpsert(&upsert))
	}

	// build the batches - make sure EVERYTHING is as we expect - two batches, 3 upserts in the first, 2 in the second

	// batch 1: 3 upserts
	batchBytes, upsertCount, err := sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(3, upsertCount)
	expectedBatch := TagBatchPart{
		BatchGUID:  "",
		ReplaceAll: true,
		IsComplete: false,
		Upserts: []TagUpsert{
			TagUpsert{
				Value: "value_A_A_A_A_A_5",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.5"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_A_A_A_4",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_A_A_3",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.3"},
					},
				},
			},
		},
		TTLMinutes: 17,
		Sender: TagBatchPartSender{
			ServiceName:     "service-name",
			ServiceInstance: "service-instance",
			HostName:        "host-name",
		},
	}
	receivedBatch := TagBatchPart{}
	a.NoError(json.Unmarshal(batchBytes, &receivedBatch))
	a.True(expectedBatch.Equal(receivedBatch))

	// batch 2: 2 upserts
	sut.SetBatchGUID("805e4dcb-3ecd-24f3-3a35-3e926e4bded5")
	batchBytes, upsertCount, err = sut.BuildBatchRequest()
	a.NoError(err)
	a.Equal(2, upsertCount)
	expectedBatch = TagBatchPart{
		BatchGUID:  "805e4dcb-3ecd-24f3-3a35-3e926e4bded5",
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			TagUpsert{
				Value: "value_A_A_2",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.2"},
					},
				},
			},
			TagUpsert{
				Value: "value_A_1",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.1"},
					},
				},
			},
		},
		TTLMinutes: 17,
		Sender: TagBatchPartSender{
			ServiceName:     "service-name",
			ServiceInstance: "service-instance",
			HostName:        "host-name",
		},
	}
	receivedBatch = TagBatchPart{}
	a.NoError(json.Unmarshal(batchBytes, &receivedBatch))
	a.True(expectedBatch.Equal(receivedBatch))

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
						Direction:   "src",
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
						Direction:   "src",
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
			batchBytes, upsertCount, err := sut.BuildBatchRequest()
			a.NoError(err)
			a.NotNil(batchBytes)
			a.Equal(batchSizeBytes, len(batchBytes))
			a.True(len(batchBytes) <= maxSize)
			a.Equal(bigCount+smallCount, upsertCount)

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
		batchBytes, upsertCount, err := sut.BuildBatchRequest()
		a.Nil(batchBytes)
		a.Error(err)
		a.Equal("Only first batch may be sent without batch GUID", err.Error())
		a.Zero(upsertCount)

		// set the GUID and try again
		sut.SetBatchGUID("705e4dcb-3ecd-24f3-3a35-3e926e4bded5")

		verifyBatch(1, 2, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 12066, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11740, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11740, false, 13)
		verifyBatch(1, 0, "705e4dcb-3ecd-24f3-3a35-3e926e4bded5", 11739, true, 13)

		// make sure that the batch builder knows it's done
		batchBytes, upsertCount, err = sut.BuildBatchRequest()
		a.Nil(batchBytes)
		a.NoError(err)
		a.Zero(upsertCount)
	}

	// run the test twice to make sure we can reuse SUT
	runTest()
	sut.Reset(maxSize, true, 13)
	runTest()
}
