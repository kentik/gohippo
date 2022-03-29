package hippo

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// helper function to assert that the headers are correct, gzip uncompresses the request,
// and returns the JSON
func getJSON(a *require.Assertions, r *http.Request) []byte {
	// verify Content-Encoding
	a.Equal(1, len(r.Header["Content-Encoding"]))
	a.Equal("gzip", r.Header["Content-Encoding"][0])

	// verify Content-Type
	a.Equal(1, len(r.Header["Content-Type"]))
	a.Equal("application/json", r.Header["Content-Type"][0])

	gzippedPayload, err := ioutil.ReadAll(r.Body)
	a.NoError(err)

	return gzipUncompress(a, gzippedPayload)
}

// Test sending a small batch through a fake server
func TestSinglePartBatch_Success(t *testing.T) {
	a := require.New(t)

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

		jsonPayload := getJSON(a, r)

		// verify the expected request
		expectedRequest := `{"guid":"","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test1","criteria":[{"direction":"src","addr":["1.2.3.4"]}]}],"complete":true}`
		a.Equal(expectedRequest, string(jsonPayload))

		// write the canned response
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// nolint
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
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(context.Background(), url, &batch)
	a.NoError(err)
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(1, response.PartsSent)
	a.Equal(1, response.UpsertsSent)
	a.Equal(1, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)
}

// Test sending a small batch through a fake server - not so concerned here with the structure of upserts
func TestSinglePartBatch_Error(t *testing.T) {
	a := require.New(t)

	serviceCalled := false

	cannedResponse := &APIServerResponse{
		GUID:    "",
		Message: "",
		Error:   "Internal error processing request - please re-submit this batch part, or the entire batch",
	}

	// setup test server
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()
	mux.HandleFunc("/kentik/server/url", func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		serviceCalled = true

		jsonPayload := getJSON(a, r)

		// verify the expected request
		expectedRequest := `{"guid":"","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test1","criteria":[{"direction":"src","addr":["1.2.3.4"]}]}],"complete":true}`
		a.Equal(expectedRequest, string(jsonPayload))

		// write the canned response
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// nolint
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
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(context.Background(), url, &batch)
	a.Error(err)
	expectedError := `API response contained an error - [Batch GUID: ; Progress: 0 parts sent, 0/1 upserts, 0/0 deletes] - server message: ; server error: Internal error processing request - please re-submit this batch part, or the entire batch`
	a.Equal(expectedError, err.Error())

	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(0, response.PartsSent)
	a.Equal(0, response.UpsertsSent)
	a.Equal(1, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal("", response.BatchGUID)
}

// Test sending a small batch through a fake server - not so concerned here with the structure of upserts
func TestSinglePartBatch_MissingGUID(t *testing.T) {
	a := require.New(t)

	serviceCalled := false

	cannedResponse := &APIServerResponse{
		GUID:    "",
		Message: "",
		Error:   "",
	}

	// setup test server
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()
	mux.HandleFunc("/kentik/server/url", func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		serviceCalled = true
		jsonPayload := getJSON(a, r)

		// verify the expected request
		expectedRequest := `{"guid":"","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test1","criteria":[{"direction":"src","addr":["1.2.3.4"]}]}],"complete":true}`
		a.Equal(expectedRequest, string(jsonPayload))

		// write the canned response
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// nolint
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
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(context.Background(), url, &batch)
	a.Error(err)
	expectedError := `API response did not include a GUID for subsequent batches - [Batch GUID: ; Progress: 0 parts sent, 0/1 upserts, 0/0 deletes] - server message: ; server error: `
	a.Equal(expectedError, err.Error())

	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(0, response.PartsSent)
	a.Equal(0, response.UpsertsSent)
	a.Equal(1, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal("", response.BatchGUID)
}

// Test sending a multiple-batch through a fake server - not so concerned here with the structure of upserts
func TestMultiPartBatch_Success(t *testing.T) {
	a := require.New(t)

	serviceCalled := false

	// same response both times
	cannedResponse := &APIServerResponse{
		GUID:    "c8285742-f7a4-4870-933d-665b15c31eda",
		Message: "success",
		Error:   "",
	}

	// expecting 3 requests
	expectedRequests := []string{
		// first request has no GUID, and has complete=false
		`{"guid":"","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test5","criteria":[{"direction":"src","addr":["5.2.3.4"]}]},{"value":"test4","criteria":[{"direction":"src","addr":["4.2.3.4"]}]}],"complete":false}`,

		// second request has the GUID, and has complete=false
		`{"guid":"c8285742-f7a4-4870-933d-665b15c31eda","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test3","criteria":[{"direction":"src","addr":["3.2.3.4"]}]},{"value":"test2","criteria":[{"direction":"src","addr":["2.2.3.4"]}]}],"complete":false}`,

		// third request has the GUID, and complete=true
		`{"guid":"c8285742-f7a4-4870-933d-665b15c31eda","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test1","criteria":[{"direction":"src","addr":["1.2.3.4"]}]}],"complete":true}`,
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

		jsonPayload := getJSON(a, r)

		// keep track of the requests received
		receivedRequests = append(receivedRequests, string(jsonPayload))

		// write the canned response - same for both batches
		responseBytes, err := json.Marshal(cannedResponse)
		a.NoError(err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// nolint
		w.Write(responseBytes)
	})

	sut := NewHippo("agent", "email", "token")
	sut.SetSenderInfo("my-service", "service-instance-1", "my-host-name")

	// force the client to use small batches
	sut.OutgoingRequestSize = 380

	// build the request
	batch := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			{
				Value: "test1",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			{
				Value: "test2",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"2.2.3.4"},
					},
				},
			},
			{
				Value: "test3",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"3.2.3.4"},
					},
				},
			},
			{
				Value: "test4",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"4.2.3.4"},
					},
				},
			},
			{
				Value: "test5",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"5.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(context.Background(), url, &batch)
	a.NoError(err)
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(3, response.PartsSent)
	a.Equal(5, response.UpsertsSent)
	a.Equal(5, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)

	a.Equal(len(expectedRequests), len(receivedRequests))
	a.Equal(expectedRequests[0], receivedRequests[0])
	a.Equal(expectedRequests[1], receivedRequests[1])
	a.Equal(expectedRequests[2], receivedRequests[2])

	// make sure all the upserts are found
	foundValues := make(map[string]bool)
	for i := 0; i < 3; i++ {
		batch := TagBatchPart{}
		a.NoError(json.Unmarshal([]byte(receivedRequests[i]), &batch))
		for _, upsert := range batch.Upserts {
			a.False(foundValues[upsert.Value])
			foundValues[upsert.Value] = true
		}
	}
	a.Equal(5, len(foundValues))
	for i := 1; i <= 5; i++ {
		a.True(foundValues[fmt.Sprintf("test%d", i)])
	}
}

// Test sending a multiple-batch through a fake server - not so concerned here with the structure of upserts - partial success
func TestMultiPartBatch_PartialSuccess(t *testing.T) {
	a := require.New(t)

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
		`{"guid":"","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test5","criteria":[{"direction":"src","addr":["5.2.3.4"]}]},{"value":"test4","criteria":[{"direction":"src","addr":["4.2.3.4"]}]}],"complete":false}`,

		// second request has the GUID, and also has complete=false, since there should be 3 parts, but we only send 2
		`{"guid":"c8285742-f7a4-4870-933d-665b15c31eda","replace_all":true,"ttl_minutes":0,"sender":{"service_name":"my-service","service_instance":"service-instance-1","host_name":"my-host-name"},"upserts":[{"value":"test3","criteria":[{"direction":"src","addr":["3.2.3.4"]}]},{"value":"test2","criteria":[{"direction":"src","addr":["2.2.3.4"]}]}],"complete":false}`,
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

		jsonPayload := getJSON(a, r)

		// keep track of the requests received
		receivedRequests = append(receivedRequests, string(jsonPayload))
		requestCount++

		// first request is success, second is failure
		if requestCount == 1 {
			// write the canned response - same for both batches
			responseBytes, err := json.Marshal(cannedResponse)
			a.NoError(err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			// nolint
			w.Write(responseBytes)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			// nolint
			w.Write([]byte("server error occurred"))
		}
	})

	sut := NewHippo("agent", "email", "token")
	sut.SetSenderInfo("my-service", "service-instance-1", "my-host-name")

	// force the client to use small batches
	sut.OutgoingRequestSize = 380

	// build the request
	batch := TagBatchPart{
		ReplaceAll: true,
		IsComplete: true,
		Upserts: []TagUpsert{
			{
				Value: "test1",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"1.2.3.4"},
					},
				},
			},
			{
				Value: "test2",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"2.2.3.4"},
					},
				},
			},
			{
				Value: "test3",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"3.2.3.4"},
					},
				},
			},
			{
				Value: "test4",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"4.2.3.4"},
					},
				},
			},
			{
				Value: "test5",
				Criteria: []TagCriteria{
					{
						Direction:   "src",
						IPAddresses: []string{"5.2.3.4"},
					},
				},
			},
		},
		TTLMinutes: 0,
	}

	url := fmt.Sprintf("%s/kentik/server/url", ts.URL)
	response, err := sut.SendBatch(context.Background(), url, &batch)
	a.Error(err)
	expectedErrorStr := fmt.Sprintf(`Error POSTing populators to %s/kentik/server/url (232 bytes) - [Batch GUID: c8285742-f7a4-4870-933d-665b15c31eda; Progress: 1 parts sent, 2/5 upserts, 0/0 deletes] - underlying error: http error 500: server error occurred`, ts.URL)
	a.Equal(expectedErrorStr, err.Error())
	a.NotNil(response)

	// make sure the fake web service was hit
	a.True(serviceCalled)

	// verify the response
	a.Equal(1, response.PartsSent)
	a.Equal(2, response.UpsertsSent)
	a.Equal(5, response.UpsertsTotal)
	a.Equal(0, response.DeletesSent)
	a.Equal(0, response.DeletesTotal)
	a.Equal(cannedResponse.GUID, response.BatchGUID)

	a.Equal(2, len(receivedRequests)) // wanted 3, but only got 2
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
	require.True(t, compactTagBatchPart.ReplaceAll)
	require.True(t, compactTagBatchPart.IsComplete)
	require.Equal(t, 1, len(compactTagBatchPart.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	require.Equal(t, "My device", compactTagBatchPart.Upserts[0].Value)
	require.Equal(t, 2, len(compactTagBatchPart.Upserts[0].Criteria))
	require.Equal(t, "dst", compactTagBatchPart.Upserts[0].Criteria[0].Direction)
	require.Equal(t, "dst", compactTagBatchPart.Upserts[0].Criteria[1].Direction)
	require.Equal(t, 1, len(compactTagBatchPart.Upserts[0].Criteria[0].DeviceNameRegexes))
	require.Equal(t, 1, len(compactTagBatchPart.Upserts[0].Criteria[1].DeviceNameRegexes))
	require.Equal(t, "foo", compactTagBatchPart.Upserts[0].Criteria[0].DeviceNameRegexes[0])
	require.Equal(t, "bar", compactTagBatchPart.Upserts[0].Criteria[1].DeviceNameRegexes[0])
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
	require.True(t, compactTagBatchPart.ReplaceAll)
	require.True(t, compactTagBatchPart.IsComplete)
	require.Equal(t, 2, len(compactTagBatchPart.Upserts))

	// we don't force the case on the value, it just happens that we take the last case seen
	require.True(t, ("my device 1" == compactTagBatchPart.Upserts[0].Value && "My device 2" == compactTagBatchPart.Upserts[1].Value) ||
		("my device 1" == compactTagBatchPart.Upserts[1].Value && "My device 2" == compactTagBatchPart.Upserts[0].Value))
}

func TestFlexStringCriteriaEncoding(t *testing.T) {
	require := require.New(t)

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
	require.NoError(err)

	actual, err := json.MarshalIndent(&rule, "", "  ")
	require.NoError(err)

	require.Equal(string(expect), string(actual))
}

// build a list of IP addresses
func buildIPAddresses(count int) []string {
	ret := make([]string, 0, count)
	for a := 1; a < 255; a++ {
		for b := 1; b < 255; b++ {
			for c := 1; c < 255; c++ {
				for d := 1; d < 255; d++ {
					ret = append(ret, fmt.Sprintf("%d.%d.%d.%d", a, b, c, d))
					if len(ret) >= count {
						return ret
					}
				}
			}
		}
	}
	return ret
}

func gzipUncompress(a *require.Assertions, data []byte) []byte {
	b := bytes.NewBuffer(data)

	r, err := gzip.NewReader(b)
	a.NoError(err)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	a.NoError(err)

	return resB.Bytes()
}
