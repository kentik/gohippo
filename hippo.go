package hippo

import (
	"bytes"
	"context"
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
