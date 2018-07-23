package hippo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Replace  bool     `json:"replace_all"`
	Complete bool     `json:"complete"`
	Upserts  []Upsert `json:"upserts,omitempty"`
	Deletes  []Delete `json:"deletes,omitempty"`
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
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = fmt.Errorf("http error %d: %s", resp.StatusCode, buf)
		}
		return nil, err
	}
	return buf, nil
}

func (c *Client) EncodeReq(r *Req) ([]byte, error) {
	if b, err := json.Marshal(r); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}
