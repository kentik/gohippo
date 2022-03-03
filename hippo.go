package hippo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type SendBatchResult struct {
	PartsSent    int
	PartsTotal   int
	UpsertsSent  int
	UpsertsTotal int
	DeletesSent  int
	DeletesTotal int
	BatchGUID    string
}

func (r *SendBatchResult) String() string {
	return fmt.Sprintf("Batch GUID: %s; Progress: %d/%d parts, %d/%d upserts, %d/%d deletes", r.BatchGUID, r.PartsSent, r.PartsTotal, r.UpsertsSent, r.UpsertsTotal, r.DeletesSent, r.DeletesTotal)
}

const (
	MAX_TAG_LEN            = 128
	DEFAULT_MAX_HIPPO_SIZE = 3000000 // split requests up which are bigger than this (3MB ish)
)

type Client struct {
	http                *http.Client
	transport           *http.Transport
	UsrAgent            string
	UsrEmail            string
	UsrToken            string
	OutgoingRequestSize int

	Sender TagBatchPartSender // optional - to help track batch origin
	lock   sync.RWMutex
}

const (
	FlexStringActionExact  = "exact"
	FlexStringActionPrefix = "prefix"
)

type CustomDimension struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsBulk      bool   `json:"is_bulk"`
	IsInternal  bool   `json:"internal"`
}

type CustomDimensionList struct {
	Dimensions []*CustomDimension `json:"customDimensions"`
}

func NewHippo(agent string, email string, token string) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
	}

	return &Client{
		http:                client,
		transport:           transport,
		UsrAgent:            agent,
		UsrEmail:            email,
		UsrToken:            token,
		OutgoingRequestSize: DEFAULT_MAX_HIPPO_SIZE,
	}
}

// SetSenderInfo sets optional metadata about the service sending batches
func (c *Client) SetSenderInfo(serviceName string, serviceInstance string, hostName string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Sender = TagBatchPartSender{
		ServiceName:     serviceName,
		ServiceInstance: serviceInstance,
		HostName:        hostName,
	}
}

func (c *Client) SetProxy(url *url.URL) {
	c.transport.Proxy = http.ProxyURL(url)
}

func (c *Client) NewTagBatchPartRequest(method string, url string, data []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UsrAgent)
	req.Header.Set("X-CH-Auth-Email", c.UsrEmail)
	req.Header.Set("X-CH-Auth-API-Token", c.UsrToken)
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) ([]byte, error) {
	req = req.WithContext(ctx)
	resp, err := c.http.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil || (resp.StatusCode >= 300) {
		if err == nil {
			err = fmt.Errorf("http error %d: %s", resp.StatusCode, buf)
		}
		return nil, err
	}
	return buf, nil
}

// SendBatchPart sends a batch to the server, in multiple requests, if necessary.
// - may return non-nil SendBatchResponse on error, it'll report how far we got
// - error will contain SendBatchResponse info
func (c *Client) SendBatch(ctx context.Context, url string, batch *TagBatchPart) (*SendBatchResult, error) {
	c.lock.RLock()
	sender := c.Sender
	c.lock.RUnlock()

	// compact the batch, grouping the same values together
	batch = compactTagBatchPart(*batch)
	batch.Sender = sender

	batchParts, err := c.split(batch)
	if err != nil {
		return nil, err
	}

	ret := &SendBatchResult{
		PartsTotal:   len(batchParts),
		UpsertsTotal: len(batch.Upserts),
		DeletesTotal: len(batch.Deletes),
		BatchGUID:    "", // not known until we send the first part
	}

	for i, batchPart := range batchParts {
		batchPart.BatchGUID = ret.BatchGUID

		requestBytes, err := json.Marshal(batchPart)
		if err != nil {
			return ret, fmt.Errorf("Error JSON marshalling request batch - [%s] - underlying error: %s", ret, err)
		}

		req, err := c.NewTagBatchPartRequest("POST", url, requestBytes)
		if err != nil {
			return ret, fmt.Errorf("Error building request to %s - [%s] - underlying error: %s", url, ret, err)
		}
		responseBytes, err := c.Do(ctx, req)
		if err != nil {
			return ret, fmt.Errorf("Error POSTing populators to %s - [%s] - underlying error: %s", url, ret, err)
		}

		if i == 0 {
			// first response returns the batch GUID, which we need to include in subsequent batches
			apiResponse := APIServerResponse{}
			if err := json.Unmarshal(responseBytes, &apiResponse); err != nil {
				return ret, fmt.Errorf("Error unmarshalling API batch response - [%s] - underlying error: %s", ret, err)
			}
			if apiResponse.GUID == "" {
				return ret, fmt.Errorf("API response did not include a GUID for subsequent batches - [%s] - underlying error: %s", ret, apiResponse.Error)
			}
			ret.BatchGUID = apiResponse.GUID
		}

		// update response
		ret.PartsSent++
		ret.UpsertsSent += len(batchPart.Upserts)
		ret.DeletesSent += len(batchPart.Deletes)

		// slow down the HTTP batches a bit to avoid rate limiting
		time.Sleep(time.Second)
	}

	return ret, nil
}

// split a batch into parts small enough for sending through Kentik API
func (c *Client) split(rFull *TagBatchPart) ([]TagBatchPart, error) {
	encode := func(r *TagBatchPart) ([]byte, error) {
		if b, err := json.Marshal(r); err != nil {
			return nil, fmt.Errorf("Error JSON-marshalling batch: %s", err)
		} else {
			return b, nil
		}
	}

	serializedBytes, err := encode(rFull)
	if err != nil {
		return nil, err
	}

	outgoingRequestSize := c.OutgoingRequestSize
	if outgoingRequestSize <= 0 {
		outgoingRequestSize = DEFAULT_MAX_HIPPO_SIZE
	}

	if len(serializedBytes) < outgoingRequestSize {
		return []TagBatchPart{*rFull}, nil
	}

	// Here, have to split this request into a group
	parts := (len(serializedBytes) / outgoingRequestSize) + 1
	upsertPerPart := len(rFull.Upserts) / parts
	ret := make([]TagBatchPart, parts)
	lastUp := 0

	for i := 0; i < parts-1; i++ {
		ret[i] = TagBatchPart{
			ReplaceAll: rFull.ReplaceAll,
			IsComplete: false,
			TTLMinutes: rFull.TTLMinutes,
			Upserts:    rFull.Upserts[lastUp : lastUp+upsertPerPart],
			Sender:     rFull.Sender,
		}

		lastUp += upsertPerPart
	}

	// Last one has to be handled seperately
	ret[parts-1] = TagBatchPart{
		ReplaceAll: rFull.ReplaceAll,
		IsComplete: true,
		TTLMinutes: rFull.TTLMinutes,
		Upserts:    rFull.Upserts[lastUp:],
		Sender:     rFull.Sender,
	}

	return ret, nil
}

// Compact a request down to combine criteria with the same values, returning a new request.
// - returned struct shouldn't be modified, because it shares slices with the original
func compactTagBatchPart(rFull TagBatchPart) *TagBatchPart {
	rulesByLowerValue := make(map[string][]TagCriteria)
	valByLowerVal := make(map[string]string)

	for _, upsert := range rFull.Upserts {
		valLower := strings.ToLower(upsert.Value)
		valByLowerVal[valLower] = upsert.Value
		if _, found := rulesByLowerValue[valLower]; !found {
			rulesByLowerValue[valLower] = make([]TagCriteria, 0, len(upsert.Criteria))
		}
		for _, rule := range upsert.Criteria {
			rulesByLowerValue[valLower] = append(rulesByLowerValue[valLower], rule)
		}
	}

	// re-build the upserts collection
	// - start with a copied instance, which shares the underlying slices
	// - then replace the Upserts slice
	ret := rFull
	ret.Upserts = make([]TagUpsert, 0, len(rulesByLowerValue))
	for valLower, rules := range rulesByLowerValue {
		ret.Upserts = append(ret.Upserts, TagUpsert{
			Value:    valByLowerVal[valLower],
			Criteria: rules,
		})
	}

	// sort upserts by value - for testability and possibly to help make errors more understandable on server side
	sort.SliceStable(ret.Upserts, func(i, j int) bool {
		return ret.Upserts[i].Value < ret.Upserts[j].Value
	})

	return &ret
}

// Create any dimensions which are not present for the given company.
func (c *Client) EnsureDimensions(ctx context.Context, apiHost string, required map[string]string) (int, error) {
	var currentSet CustomDimensionList
	found := map[string]bool{}
	done := 0

	for col, _ := range required {
		found[col] = false
	}

	url := fmt.Sprintf("%s/api/internal/customdimensions", apiHost)
	if req, err := c.NewTagBatchPartRequest("GET", url, nil); err != nil {
		return done, err
	} else {
		if res, err := c.Do(ctx, req); err != nil {
			return done, err
		} else {
			if err := json.Unmarshal(res, &currentSet); err != nil {
				return done, err
			} else {
				for _, dim := range currentSet.Dimensions {
					if _, ok := found[dim.Name]; ok {
						found[dim.Name] = true
					}
				}
			}
		}
	}

	// Now, try to make any dimensions not found
	for col, present := range found {
		if !present {
			cd := CustomDimension{
				DisplayName: required[col],
				Name:        col,
				Type:        "string",
				IsBulk:      true,
				IsInternal:  strings.HasPrefix(col, "kt_"),
			}
			if b, err := json.Marshal(cd); err != nil {
				return done, err
			} else {
				url := fmt.Sprintf("%s/api/internal/customdimension", apiHost)
				if req, err := c.NewTagBatchPartRequest("POST", url, b); err != nil {
					return done, err
				} else {
					if _, err := c.Do(ctx, req); err != nil {
						if !strings.Contains(err.Error(), "already in use") {
							//fmt.Printf("Warn, ensuring dimension failed: %s\n", err)
							return done, err
						}
						// The column already exists. This can happen for columns that don't have a 'c_' prefix, such as 'kt_' columns.
						// We get into this situation because our API call to get existing columns doesn't return these columns,
						// so we're here trying to create them.
					} else {
						done++
					}
				}
			}
		}
	}
	return done, nil
}

func TruncateStringForMaxTagLen(str string) string {
	if len(str) > MAX_TAG_LEN {
		return str[0:MAX_TAG_LEN]
	}
	return str
}
