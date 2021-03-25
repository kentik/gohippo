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
	"strings"
	"sync"
	"time"
)

const (
	MAX_TAG_LEN    = 128
	MAX_HIPPO_SIZE = 3000000 // split requests up which are bigger than this (3MB ish)
)

type Client struct {
	http      *http.Client
	transport *http.Transport
	UsrAgent  string
	UsrEmail  string
	UsrToken  string

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
		http:      client,
		transport: transport,
		UsrAgent:  agent,
		UsrEmail:  email,
		UsrToken:  token,
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

func (c *Client) EncodeTagBatchPart(rFull *TagBatchPart) ([][]byte, int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// swap out our local copy with a compacted one that ensures all criteria are grouped by value
	tmp := compactTagBatchPart(*rFull)
	tmp.Sender = c.Sender
	rFull = &tmp

	encode := func(r *TagBatchPart) ([]byte, error) {
		if b, err := json.Marshal(r); err != nil {
			return nil, err
		} else {
			return b, nil
		}
	}

	// If the size is small enough, just return here. Assume default case
	init, err := encode(rFull)
	if err != nil {
		return nil, 0, err
	} else if len(init) < MAX_HIPPO_SIZE {
		return [][]byte{init}, len(rFull.Upserts), nil
	}

	// Here, have to split this request into a group
	parts := (len(init) / MAX_HIPPO_SIZE) + 1
	upsertPerPart := len(rFull.Upserts) / parts
	reqArrary := make([][]byte, parts)
	lastUp := 0
	numUpserts := 0

	for i := 0; i < parts-1; i++ {
		rTmp := &TagBatchPart{
			ReplaceAll: rFull.ReplaceAll,
			IsComplete: false,
			TTLMinutes: rFull.TTLMinutes,
			Upserts:    rFull.Upserts[lastUp : lastUp+upsertPerPart],
			Sender:     c.Sender,
		}

		next, err := encode(rTmp)
		if err != nil {
			return nil, 0, err
		} else {
			reqArrary[i] = next
		}
		lastUp += upsertPerPart
		numUpserts += len(rTmp.Upserts)
	}

	// Last one has to be handled seperately
	rLast := &TagBatchPart{
		ReplaceAll: rFull.ReplaceAll,
		IsComplete: true,
		TTLMinutes: rFull.TTLMinutes,
		Upserts:    rFull.Upserts[lastUp:],
		Sender:     c.Sender,
	}

	numUpserts += len(rLast.Upserts)
	next, err := encode(rLast)
	if err != nil {
		return nil, 0, err
	} else {
		reqArrary[parts-1] = next
	}

	return reqArrary, numUpserts, nil
}

// Compact a request down to combine criteria with the same values, returning a new request.
// - returned struct shouldn't be modified, because it shares slices with the original
func compactTagBatchPart(rFull TagBatchPart) TagBatchPart {
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
	return ret
}

// Create any dimensions which are not present for the given company.
func (c *Client) EnsureDimensions(apiHost string, required map[string]string) (int, error) {
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
		if res, err := c.Do(context.Background(), req); err != nil {
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
					if _, err := c.Do(context.Background(), req); err != nil {
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
