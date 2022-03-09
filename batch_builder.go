package hippo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// BatchBuilder is responsible for building a batch that will serialize within a desired size if possible.
// It'll fit big and small populators together, but doesn't try to do so optimally.
type BatchBuilder struct {
	desiredSize       int
	buf               *bytes.Buffer
	serializedUpserts [][]byte
	batchGUID         string
	replaceAll        bool
	ttlMinutes        uint32
	builtBatchesCount int
	hasClosedBatch    bool
	sender            TagBatchPartSender // optional - to help track batch origin
}

// NewBatchBuilder builds a new BatchBuilder
func NewBatchBuilder(desiredSize int, replaceAll bool, ttlMinutes uint32) *BatchBuilder {
	return &BatchBuilder{
		desiredSize:       desiredSize,
		serializedUpserts: make([][]byte, 0),
		replaceAll:        replaceAll,
		ttlMinutes:        ttlMinutes,
		buf:               bytes.NewBuffer(make([]byte, 0, desiredSize)),
	}
}

// Reset resets the batch builder to build a new batch, reusing the underlying buffer.
func (b *BatchBuilder) Reset(desiredSize int, replaceAll bool, ttlMinutes uint32) {
	b.desiredSize = desiredSize
	b.serializedUpserts = b.serializedUpserts[:0]
	b.batchGUID = ""
	b.replaceAll = replaceAll
	b.ttlMinutes = ttlMinutes
	b.builtBatchesCount = 0
	b.hasClosedBatch = false
}

// SetSenderInfo sets optional metadata about the service sending batches
func (b *BatchBuilder) SetSenderInfo(sender TagBatchPartSender) {
	b.sender = sender
}

func (b *BatchBuilder) isSenderInfoSet() bool {
	return b.sender.ServiceName != "" || b.sender.ServiceInstance != "" || b.sender.HostName != ""
}

// AddUpsert attempts to add the input upsert into the batch.
// Make sure to add all upserts to the batch before calling BuildBatch()
func (b *BatchBuilder) AddUpsert(upsert *TagUpsert) error {
	ser, err := json.Marshal(upsert)
	if err != nil {
		return fmt.Errorf("Error serializing TagUpsert: %s", err)
	}
	b.serializedUpserts = append(b.serializedUpserts, ser)
	return nil
}

// SetBatchGUID sets the GUID that was returned after submitting the first part to the server.
func (b *BatchBuilder) SetBatchGUID(guid string) {
	b.batchGUID = guid
}

// BuildBatch builds and returns a serialized batch.
// Once this method is called, don't make any further calls to AddUpsert.
// - returns serialized batch and upsert count
func (b *BatchBuilder) BuildBatch() ([]byte, int, error) {
	if len(b.serializedUpserts) == 0 && (b.hasClosedBatch || !b.replaceAll) {
		// nothing to do
		return nil, 0, nil
	}

	if b.batchGUID == "" && b.builtBatchesCount > 0 {
		return nil, 0, fmt.Errorf("Only first batch may be sent without batch GUID")
	}

	b.buf.Reset()

	// sort upserts by serialized size
	sort.SliceStable(b.serializedUpserts, func(i int, j int) bool {
		return len(b.serializedUpserts[i]) < len(b.serializedUpserts[j])
	})

	// guid
	if _, err := b.buf.WriteString(`{"guid":"`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}
	if _, err := b.buf.WriteString(b.batchGUID); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	// replace_all
	if _, err := b.buf.WriteString(`","replace_all":`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}
	if _, err := b.buf.WriteString(boolString(b.replaceAll)); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	// ttl_minutes
	if _, err := b.buf.WriteString(`,"ttl_minutes":`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}
	if _, err := b.buf.WriteString(fmt.Sprintf("%d", b.ttlMinutes)); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	// service info, if set
	if b.isSenderInfoSet() {
		senderBytes, err := json.Marshal(b.sender)
		if err != nil {
			return nil, 0, fmt.Errorf("Error marshalling sender info to JSON: %s", err)
		}

		if _, err := b.buf.WriteString(`,"sender":`); err != nil {
			return nil, 0, fmt.Errorf("Error writing sender info to buffer: %s", err)
		}
		if _, err := b.buf.Write(senderBytes); err != nil {
			return nil, 0, fmt.Errorf("Error writing sender info to buffer: %s", err)
		}
	}

	// upserts start
	if _, err := b.buf.WriteString(`,"upserts":[`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	// build a batch as big as we can
	start := 0
	end := len(b.serializedUpserts) - 1

	// leave some space for batch scaffolding and `"complete":false`
	availableSpace := b.desiredSize - b.buf.Len() - 40

	upsertCount := 0
	for start <= end && availableSpace > 0 {
		// try the bigger upserts first
		if len(b.serializedUpserts[end]) <= availableSpace {
			if upsertCount > 0 {
				if _, err := b.buf.WriteString(","); err != nil {
					return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
				}
				availableSpace--
			}
			if _, err := b.buf.Write(b.serializedUpserts[end]); err != nil {
				return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
			}
			availableSpace -= len(b.serializedUpserts[end])

			b.serializedUpserts[end] = nil

			upsertCount++
			end--
			continue
		}

		// couldn't fit the bigger one - try smaller
		if len(b.serializedUpserts[start]) <= availableSpace {
			if upsertCount > 0 {
				if _, err := b.buf.WriteString(","); err != nil {
					return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
				}
				availableSpace--
			}
			if _, err := b.buf.Write(b.serializedUpserts[start]); err != nil {
				return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
			}
			availableSpace -= len(b.serializedUpserts[start])

			b.serializedUpserts[start] = nil

			upsertCount++
			start++
			continue
		}

		// smaller wouldn't fit either
		if upsertCount == 0 {
			// batch is empty
			return nil, 0, fmt.Errorf("Have %d remaining upserts, but could not fit any into the batch. The smallest one is %d bytes", len(b.serializedUpserts), len(b.serializedUpserts[start]))
		}
		break
	}

	if _, err := b.buf.WriteString(`]`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	// is_complete
	if _, err := b.buf.WriteString(`,"complete":`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}
	if _, err := b.buf.WriteString(boolString(start > end)); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}
	b.hasClosedBatch = start > end

	if _, err := b.buf.WriteString(`}`); err != nil {
		return nil, 0, fmt.Errorf("Error writing string to buffer: %s", err)
	}

	b.serializedUpserts = b.serializedUpserts[start : end+1]

	b.builtBatchesCount++
	return b.buf.Bytes(), upsertCount, nil
}

func boolString(val bool) string {
	if val {
		return "true"
	}
	return "false"
}
