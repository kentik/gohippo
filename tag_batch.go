package hippo

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// compiled regular expression - for tags - custom dimensions can have whatever they want
var _validTagValueRegexp *regexp.Regexp

// init function for this file - evaluated in init.go
func initTagBatch() {
	validTagValueRegexp, err := regexp.Compile(`^[a-zA-Z0-9_-]+$`)
	if err != nil {
		panic(fmt.Sprintf("Cannot compile validTagValueRegexp: %s", err))
	}
	_validTagValueRegexp = validTagValueRegexp
}

// NewTagBatch returns a new TagBatch
func NewTagBatch() TagBatchPart {
	return TagBatchPart{
		Upserts: make([]TagUpsert, 0),
		Deletes: make([]TagDelete, 0),
	}
}

// Validate validates the record, returning true if valid, and a customer-friendly error string if false
func (t *TagDelete) Validate() (bool, string) {
	if t.Value == "" {
		return false, "value cannot be empty"
	}
	return true, ""
}

// TagHashFromCriteriaHashes creates a tag hash for a column/value based on its criteria's hashes
func TagHashFromCriteriaHashes(hashes []string) string {
	sort.Strings(hashes)

	s := md5.New()
	for _, hash := range hashes {
		io.WriteString(s, hash)
	}
	return fmt.Sprintf("%x", s.Sum(nil))
}

// convert []int32 to []uint8, ignoring those that are out of bounds
func uint32sToUint8s(in []uint32) []uint8 {
	ret := make([]uint8, 0, len(in))
	for _, v := range in {
		if v >= 0 && v <= 255 {
			ret = append(ret, uint8(v))
		}
	}
	return ret
}

func ensureCIDRs(addresses []string) []string {
	ret := make([]string, len(addresses))
	for i, address := range addresses {
		parsedAddr, err := ensureCIDR(address)
		if err == nil {
			ret[i] = parsedAddr
		} else {
			// this is an invalid address, but we gotta keep it - role of this function is to add CIDR, not validate
			ret[i] = address
		}
	}
	return ret
}

func ensureCIDR(address string) (string, error) {
	var err error

	// see if there's a CIDR
	parts := strings.Split(address, "/")
	cidr := -1 // default needs to be -1 to handle /0
	if len(parts) == 2 {
		c, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return "", fmt.Errorf("couldn't parse CIDR to int: %s", err)
		}
		if c > 128 {
			return "", fmt.Errorf("Invalid CIDR: %d", c)
		}
		cidr = int(c)
	}

	// try parsing as IPv4 - force CIDR at the end
	v4AddrStr := address
	if cidr == -1 {
		// no CIDR specified - tack on /32
		v4AddrStr = fmt.Sprintf("%s/32", address)
	}
	_, ipNet, err := net.ParseCIDR(v4AddrStr)
	if err == nil {
		// parses
		if v4Addr := ipNet.IP.To4(); v4Addr != nil {
			// valid v4
			return v4AddrStr, nil
		}
		// not valid v4
	}

	// try parsing as IPv6
	v6AddrStr := address
	if cidr == -1 {
		// no CIDR specified - tack on /128
		v6AddrStr = fmt.Sprintf("%s/128", address)
	}
	_, ipNet, err = net.ParseCIDR(v6AddrStr)
	if err == nil {
		// parses
		if v6Addr := ipNet.IP.To16(); v6Addr != nil {
			// valid v6
			return v6AddrStr, nil
		}
		// not valid v6
	}

	return "", fmt.Errorf("couldn't parse either v4 or v6 address")
}
