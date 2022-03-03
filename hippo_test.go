package hippo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test sending a small batch through a fake server - not so concerned here with the structure of upserts
func TestSinglePartBatch_Success(t *testing.T) {
	a := assert.New(t)

	serviceCalled := false

	cannedResponse := &APIServerResponse{
		GUID:    "c8285742-f7a4-4870-933d-665b15c31eda",
		Message: "success",
		Error:   "",
	}

	// setup test server
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()
	mux.HandleFunc("/kentik/server/url", func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		serviceCalled = true

		jsonPayload, err := ioutil.ReadAll(r.Body)

		// verify the expected request
		expectedRequest := `{"guid":"","replace_all":true,"complete":true,"upserts":[{"value":"test1","criteria":[{"direction":"asc","addr":["1.2.3.4"]}]}],"deletes":null,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"}}`
		a.Equal(expectedRequest, string(jsonPayload))

		// write the canned response
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(responseBytes)
	})

	sut := NewHippo("agent", "email", "token")
	sut.SetSenderInfo("my-service", "service-instance-1", "my-host-name")

	// build the request
	batch := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			{
				Value: "test1",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(url, &batch)
	a.NoError(err)
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(1, response.PartsSent)
	a.Equal(1, response.PartsTotal)
	a.Equal(1, response.UpsertsSent)
	a.Equal(1, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)

}

// Test sending a multiple-batch through a fake server - not so concerned here with the structure of upserts
func TestMultiPartBatch_Success(t *testing.T) {
	a := assert.New(t)

	serviceCalled := false

	// same response both times
	cannedResponse := &APIServerResponse{
		GUID:    "c8285742-f7a4-4870-933d-665b15c31eda",
		Message: "success",
		Error:   "",
	}

	// expecting 2 requests
	expectedRequests := []string{
		// first request has no GUID, and has complete=false
		`{"guid":"","replace_all":true,"complete":false,"upserts":[{"value":"test1","criteria":[{"direction":"asc","addr":["1.2.3.4"]}]},{"value":"test2","criteria":[{"direction":"asc","addr":["2.2.3.4"]}]}],"deletes":null,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"}}`,

		// second request has the GUID, and has complete=true
		`{"guid":"c8285742-f7a4-4870-933d-665b15c31eda","replace_all":true,"complete":true,"upserts":[{"value":"test3","criteria":[{"direction":"asc","addr":["3.2.3.4"]}]},{"value":"test4","criteria":[{"direction":"asc","addr":["4.2.3.4"]}]},{"value":"test5","criteria":[{"direction":"asc","addr":["5.2.3.4"]}]}],"deletes":null,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"}}`,
	}

	lock := sync.Mutex{}
	receivedRequests := make([]string, 0)

	// setup test server
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()
	mux.HandleFunc("/kentik/server/url", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		t.Helper()
		serviceCalled = true

		// keep track of the requests received
		jsonPayload, err := ioutil.ReadAll(r.Body)
		receivedRequests = append(receivedRequests, string(jsonPayload))
		a.NoError(err)

		// write the canned response - same for both batches
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(responseBytes)
	})

	sut := NewHippo("agent", "email", "token")
	sut.SetSenderInfo("my-service", "service-instance-1", "my-host-name")

	// force the client to use small batches - request as one part is 546 bytes - make the batch size 300
	sut.OutgoingRequestSize = 300

	// build the request
	batch := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			{
				Value: "test1",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			{
				Value: "test2",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"2.2.3.4"},
					},
				},
			},
			{
				Value: "test3",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"3.2.3.4"},
					},
				},
			},
			{
				Value: "test4",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"4.2.3.4"},
					},
				},
			},
			{
				Value: "test5",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"5.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(url, &batch)
	a.NoError(err)
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)
	a.Equal(2, len(receivedRequests))

	// verify the response
	a.Equal(2, response.PartsSent)
	a.Equal(2, response.PartsTotal)
	a.Equal(5, response.UpsertsSent)
	a.Equal(5, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)

	a.Equal(expectedRequests[0], receivedRequests[0])
	a.Equal(expectedRequests[1], receivedRequests[1])
}

// Test sending a multiple-batch through a fake server - not so concerned here with the structure of upserts - partial success
func TestMultiPartBatch_PartialSuccess(t *testing.T) {
	a := assert.New(t)

	serviceCalled := false

	// same response both times
	cannedResponse := &APIServerResponse{
		GUID:    "c8285742-f7a4-4870-933d-665b15c31eda",
		Message: "success",
		Error:   "",
	}

	// expecting 2 requests
	expectedRequests := []string{
		// first request has no GUID, and has complete=false
		`{"guid":"","replace_all":true,"complete":false,"upserts":[{"value":"test1","criteria":[{"direction":"asc","addr":["1.2.3.4"]}]},{"value":"test2","criteria":[{"direction":"asc","addr":["2.2.3.4"]}]}],"deletes":null,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"}}`,

		// second request has the GUID, and has complete=true
		`{"guid":"c8285742-f7a4-4870-933d-665b15c31eda","replace_all":true,"complete":true,"upserts":[{"value":"test3","criteria":[{"direction":"asc","addr":["3.2.3.4"]}]},{"value":"test4","criteria":[{"direction":"asc","addr":["4.2.3.4"]}]},{"value":"test5","criteria":[{"direction":"asc","addr":["5.2.3.4"]}]}],"deletes":null,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"}}`,
	}

	lock := sync.Mutex{}
	receivedRequests := make([]string, 0)
	requestCount := 0

	// setup test server
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()
	mux.HandleFunc("/kentik/server/url", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		t.Helper()
		serviceCalled = true

		// keep track of the requests received
		jsonPayload, err := ioutil.ReadAll(r.Body)
		receivedRequests = append(receivedRequests, string(jsonPayload))
		a.NoError(err)
		requestCount++

		// first request is success, second is failure
		if requestCount == 1 {
			// write the canned response - same for both batches
			responseBytes, err := json.Marshal(cannedResponse)
			a.NoError(err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(responseBytes)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte("server error occurred"))
		}
	})

	sut := NewHippo("agent", "email", "token")
	sut.SetSenderInfo("my-service", "service-instance-1", "my-host-name")

	// force the client to use small batches - request as one part is 546 bytes - make the batch size 300
	sut.OutgoingRequestSize = 300

	// build the request
	batch := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			{
				Value: "test1",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			{
				Value: "test2",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"2.2.3.4"},
					},
				},
			},
			{
				Value: "test3",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"3.2.3.4"},
					},
				},
			},
			{
				Value: "test4",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"4.2.3.4"},
					},
				},
			},
			{
				Value: "test5",
				Criteria: []TagCriteria{
					{
						Direction:   "asc",
						IPAddresses: []string{"5.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(url, &batch)
	a.Error(err)
	expectedErrorStr := fmt.Sprintf(`Error POSTing populators to %s/kentik/server/url - [Batch GUID: c8285742-f7a4-4870-933d-665b15c31eda; Progress: 1/2 parts, 2/5 upserts, 0/0 deletes] - underlying error: http error 500: server error occurred`, ts.URL)
	a.Equal(expectedErrorStr, err.Error())
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)
	a.Equal(2, len(receivedRequests))

	// verify the response
	a.Equal(1, response.PartsSent)
	a.Equal(2, response.PartsTotal)
	a.Equal(2, response.UpsertsSent)
	a.Equal(5, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)

	a.Equal(expectedRequests[0], receivedRequests[0])
	a.Equal(expectedRequests[1], receivedRequests[1])
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
