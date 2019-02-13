package hippo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	MAX_TAG_LEN    = 128
	MAX_HIPPO_SIZE = 3000000 // split requests up which are bigger than this (3MB ish)
)

type Client struct {
	http     *http.Client
	UsrAgent string
	UsrEmail string
	UsrToken string
}

type Rule struct {
	Dir                string   `json:"direction,omitempty"`
	Ports              []string `json:"port,omitempty"`
	Protocols          []uint   `json:"protocol,omitempty"`
	ASNs               []string `json:"asn,omitempty"`
	VLanRanges         []string `json:"vlans,omitempty"`
	LastHopASNNames    []string `json:"lasthop_as_name,omitempty"`
	NextHopASNs        []string `json:"nexthop_asn,omitempty"`
	NextHopASNNames    []string `json:"nexthop_as_name,omitempty"`
	BGPASPaths         []string `json:"bgp_aspath,omitempty"`
	BGPCommunities     []string `json:"bgp_community,omitempty"`
	TCPFlags           uint16   `json:"tcp_flags,omitempty"`
	IPAddresses        []string `json:"addr,omitempty"`
	MACAddresses       []string `json:"mac,omitempty"`
	CountryCodes       []string `json:"country,omitempty"`
	SiteNames          []string `json:"site,omitempty"`
	DeviceTypes        []string `json:"device_type,omitempty"`
	InterfaceNames     []string `json:"interface_name,omitempty"`
	DeviceNames        []string `json:"device_name,omitempty"`
	NextHopIPAddresses []string `json:"nexthop,omitempty"`
}

type Upsert struct {
	Val   string `json:"value"`
	Rules []Rule `json:"criteria,omitempty"`
}

type Delete struct {
	Val string `json:"value"`
}

type Req struct {
	Replace    bool     `json:"replace_all"`
	Complete   bool     `json:"complete"`
	TTLMinutes int      `json:"ttl_minutes"`
	Upserts    []Upsert `json:"upserts,omitempty"`
	Deletes    []Delete `json:"deletes,omitempty"`
}

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
	c := &Client{http: http.DefaultClient, UsrAgent: agent, UsrEmail: email, UsrToken: token}
	return c
}

func (c *Client) NewRequest(method string, url string, data []byte) (*http.Request, error) {
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

func (c *Client) EncodeReq(rFull *Req) ([][]byte, int, error) {
	// swap out our local copy with a compacted one that ensures all criteria are grouped by value
	tmp := compactReq(*rFull)
	rFull = &tmp

	encode := func(r *Req) ([]byte, error) {
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
		rTmp := &Req{
			Replace:    rFull.Replace,
			Complete:   false,
			TTLMinutes: rFull.TTLMinutes,
			Upserts:    rFull.Upserts[lastUp : lastUp+upsertPerPart],
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
	rLast := &Req{
		Replace:    rFull.Replace,
		Complete:   true,
		TTLMinutes: rFull.TTLMinutes,
		Upserts:    rFull.Upserts[lastUp:],
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
func compactReq(rFull Req) Req {
	rulesByLowerValue := make(map[string][]Rule)
	valByLowerVal := make(map[string]string)

	for _, upsert := range rFull.Upserts {
		valLower := strings.ToLower(upsert.Val)
		valByLowerVal[valLower] = upsert.Val
		if _, found := rulesByLowerValue[valLower]; !found {
			rulesByLowerValue[valLower] = make([]Rule, 0, len(upsert.Rules))
		}
		for _, rule := range upsert.Rules {
			rulesByLowerValue[valLower] = append(rulesByLowerValue[valLower], rule)
		}
	}

	// re-build the upserts collection
	// - start with a copied instance, which shares the underlying slices
	// - then replace the Upserts slice
	ret := rFull
	ret.Upserts = make([]Upsert, 0, len(rulesByLowerValue))
	for valLower, rules := range rulesByLowerValue {
		ret.Upserts = append(ret.Upserts, Upsert{
			Val:   valByLowerVal[valLower],
			Rules: rules,
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
	if req, err := c.NewRequest("GET", url, nil); err != nil {
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
				if req, err := c.NewRequest("POST", url, b); err != nil {
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
