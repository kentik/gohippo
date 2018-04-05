package hippo

import (
	"context"
	"os"
	"testing"
)

func TestHippPost(t *testing.T) {
	body :=
		`{
  "replace_all": true,
  "complete": true,
  "upserts": [
    {
        "value": "foo",
        "criteria": [
          {
            "direction": "dst",
            "ports": ["4", "9", "1000", "555", "666" ,"44"],
            "protocols": [1, 2, 4, 5],
            "asns": ["12345-20000","500","10-100"],
            "last_hop_asn_names": ["asn1", "asn2"],
            "next_hop_asns": ["100-200", "35628", "60000-70000"],
            "next_hop_asn_names": ["asn2", "asn3"],
            "bgp_as_paths": ["as path 1", "as path 2"],
            "bgp_communities": ["community1", "community2"],
            "tcp_flags": 11,
            "ip_addresses": ["1.2.3.4", "3.4.5.6"],
            "mac_addresses": ["01:42:12:ae:92:bf", "03:42:1f:1e:22:bf"],
            "country_codes": ["US", "EN"],
            "site_names": ["site1", "site2"],
            "device_types": ["device_type1", "device_type2"],
            "interface_names": ["interface1", "interface2"],
            "device_names": ["device1", "device2"],
            "next_hop_ip_addresses": ["10.3.4.5", "9.6.7.8"]
          }
        ]
    },
    {
        "value": "bar",
        "criteria": [
          {
            "direction": "dst",
            "ports": ["3", "2", "3-4", "1-5", "11", "9", "1000", "555", "666", "44"],
            "protocols": [1, 2, 3, 4],
            "asns": ["12345", "23456", "100-200"],
            "last_hop_asn_names": ["asn1", "asn2"],
            "next_hop_asns": ["12345", "1-100", "23456"],
            "next_hop_asn_names": ["asn2", "asn3"],
            "bgp_as_paths": ["as path 1", "as path 2"],
            "bgp_communities": ["community1", "community2"],
            "tcp_flags": 12,
            "ip_addresses": ["1.2.3.4", "3.4.5.6"],
            "mac_addresses": ["01:42:12:ae:92:bf", "03:42:1f:1e:22:bf"],
            "country_codes": ["US", "EN"],
            "site_names": ["site1", "site2"],
            "device_types": ["device_type1", "device_type2"],
            "interface_names": ["interface1", "interface2"],
            "device_names": ["device1", "device2"],
            "next_hop_ip_addresses": ["2.9.4.5", "5.9.7.8"]
        }
      ]
    }
  ],
  "deletes": [
    {
        "value": "foo0"
    },
    {
        "value": "bar0"
    }
  ]
}`

	email := os.Getenv("API_EMAIL")
	token := os.Getenv("API_TOKEN")
    if email == "" || token == "" {
		t.Errorf("email or token not set\n")
        return
    }
	h := NewHippo("", email, token)
	if req, err := h.NewRequest("POST", "https://api.kentik.com/api/v5/tagging/tags", []byte(body)); err != nil {
		t.Errorf("req err %v\n", err)
	} else {
		if _, err := h.Do(context.Background(), req); err != nil {
			t.Errorf("resp err %v\n", err)
		}
	}
}
