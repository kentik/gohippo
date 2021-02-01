package hippo

import (
	"crypto/md5"
	"fmt"
	"github.com/cznic/sortutil"
	"io"
	"sort"
	"strconv"
	"strings"
)

// GenerateHash returns an MD5SUM hash of this TagUpsert
func (c *TagCriteria) GenerateHash() string {
	c.Normalize()

	s := md5.New()
	io.WriteString(s, c.String())
	return fmt.Sprintf("%x", s.Sum(nil))
}

func ensureStringArray(strArray []string) []string {
	if strArray == nil {
		return make([]string, 0)
	}
	return strArray
}

func ensureAndSortUint64(slice64 *[]uint64) {
	if *slice64 == nil {
		*slice64 = make([]uint64, 0)
		return
	}
	sortutil.Uint64Slice(*slice64).Sort()
}

func ensureAndSortUint32(slice32 *[]uint32) {
	if *slice32 == nil {
		*slice32 = make([]uint32, 0)
		return
	}
	sortutil.Uint32Slice(*slice32).Sort()
}

func ensureAndSortStringArray(strArray *[]string) {
	if *strArray == nil {
		*strArray = make([]string, 0)
		return
	}
	sort.Strings(*strArray)
}

func ensureAndSortFlexStringMatchArray(flexCritArray *[]FlexStringCriteria) {
	if *flexCritArray == nil {
		*flexCritArray = make([]FlexStringCriteria, 0)
		return
	}

	for i := range *flexCritArray {
		(*flexCritArray)[i].Normalize()
	}

	// sort by action, value
	sort.SliceStable(*flexCritArray, func(i, j int) bool {
		if (*flexCritArray)[i].Action != (*flexCritArray)[j].Action {
			return (*flexCritArray)[i].Action < (*flexCritArray)[j].Action
		}
		return (*flexCritArray)[i].Value < (*flexCritArray)[j].Value
	})
}

func ensureAndSortFlex32RangeArray(flexArray *[]string) {
	if *flexArray == nil {
		*flexArray = make([]string, 0)
	}
	// TODO: we could do much better here by equating "1 - 3" sorting the same as "1-3", but meh... maybe someday
	sort.Strings(*flexArray)
}

func ensureAndSortFlex64RangeArray(flexArray *[]string) {
	if *flexArray == nil {
		*flexArray = make([]string, 0)
	}
	// TODO: we could do much better here by equating "1 - 3" sorting the same as "1-3", but meh... maybe someday
	sort.Strings(*flexArray)
}

// Normalize sorts all the arrays, makes sure all fields can easily be compared
func (c *TagCriteria) Normalize() {
	asnRanges := ASNRangesSlice(ParseASNs(c.ASNRanges, nil))
	asnRanges.Sort()
	c.ASNRanges = asnRanges.ToStringArray()

	c.BGPASPaths = SanitizeBGPASPaths(ensureStringArray(c.BGPASPaths), nil)
	sort.Strings(c.BGPASPaths)

	c.BGPCommunities = SanitizeBGPCommunities(ensureStringArray(c.BGPCommunities), nil)
	sort.Strings(c.BGPCommunities)

	ensureAndSortStringArray(&c.CountryCodes)
	ensureAndSortStringArray(&c.DeviceNameRegexes)
	ensureAndSortStringArray(&c.DeviceTypeRegexes)

	c.Direction = strings.ToUpper(c.Direction)
	if c.Direction != "SRC" && c.Direction != "DST" {
		c.Direction = "EITHER"
	}

	ensureAndSortStringArray(&c.InterfaceNameRegexes)

	c.IPAddresses = ensureCIDRs(c.IPAddresses) // ensure every IP address has a CIDR
	sort.Strings(c.IPAddresses)

	ensureAndSortStringArray(&c.LastHopASNNames)

	c.MACAddresses = SanitizeMACAddresses(ensureStringArray(c.MACAddresses), nil)
	sort.Strings(c.MACAddresses)

	ensureAndSortStringArray(&c.NextHopASNNames)

	// Next-hop ASN ranges
	nextHopASNRanges := ASNRangesSlice(ParseASNs(c.NextHopASNRanges, nil))
	nextHopASNRanges.Sort()
	c.NextHopASNRanges = nextHopASNRanges.ToStringArray()

	c.NextHopIPAddresses = ensureCIDRs(c.NextHopIPAddresses) // ensure every IP address has a CIDR
	sort.Strings(c.NextHopIPAddresses)

	portRanges := PortRangesSlice(ParsePorts(c.PortRanges, nil))
	portRanges.Sort()
	c.PortRanges = portRanges.ToStringArray()

	protocols := sortutil.Uint32Slice(ParseProtocols(c.Protocols, nil))
	protocols.Sort()
	c.Protocols = protocols

	ensureAndSortStringArray(&c.SiteNameRegexes)

	vlanRanges := VLanRangesSlice(ParseVLans(c.VLanRanges, nil))
	vlanRanges.Sort()
	c.VLanRanges = vlanRanges.ToStringArray()

	// flex string columns
	ensureAndSortFlexStringMatchArray(&c.Str00)
	ensureAndSortFlexStringMatchArray(&c.Str01)
	ensureAndSortFlexStringMatchArray(&c.Str02)
	ensureAndSortFlexStringMatchArray(&c.Str03)
	ensureAndSortFlexStringMatchArray(&c.Str04)
	ensureAndSortFlexStringMatchArray(&c.Str05)
	ensureAndSortFlexStringMatchArray(&c.Str06)
	ensureAndSortFlexStringMatchArray(&c.Str07)
	ensureAndSortFlexStringMatchArray(&c.Str08)
	ensureAndSortFlexStringMatchArray(&c.Str09)
	ensureAndSortFlexStringMatchArray(&c.Str10)
	ensureAndSortFlexStringMatchArray(&c.Str11)
	ensureAndSortFlexStringMatchArray(&c.Str12)
	ensureAndSortFlexStringMatchArray(&c.Str13)
	ensureAndSortFlexStringMatchArray(&c.Str14)
	ensureAndSortFlexStringMatchArray(&c.Str15)
	ensureAndSortFlexStringMatchArray(&c.Str16)

	// flex uint64 columns
	ensureAndSortFlex64RangeArray(&c.Int6400)
	ensureAndSortFlex64RangeArray(&c.Int6401)
	ensureAndSortFlex64RangeArray(&c.Int6402)
	ensureAndSortFlex64RangeArray(&c.Int6403)
	ensureAndSortFlex64RangeArray(&c.Int6404)

	// flex app protocol column
	ensureAndSortFlex64RangeArray(&c.AppProtocol)

	// flex uint32 columns
	ensureAndSortFlex32RangeArray(&c.Int00)
	ensureAndSortFlex32RangeArray(&c.Int01)
	ensureAndSortFlex32RangeArray(&c.Int02)
	ensureAndSortFlex32RangeArray(&c.Int03)
	ensureAndSortFlex32RangeArray(&c.Int04)
	ensureAndSortFlex32RangeArray(&c.Int05)

	// flex IP addresses: ensure every IP has a CIDR, then sort
	c.Inet00 = ensureCIDRs(c.Inet00)
	sort.Strings(c.Inet00)
	c.Inet01 = ensureCIDRs(c.Inet01)
	sort.Strings(c.Inet01)
	c.Inet02 = ensureCIDRs(c.Inet02)
	sort.Strings(c.Inet02)
	c.Inet03 = ensureCIDRs(c.Inet03)
	sort.Strings(c.Inet03)
	c.Inet04 = ensureCIDRs(c.Inet04)
	sort.Strings(c.Inet04)
}

// DirectionAppliesToSource returns whether this tag applies to source flow
func (c *TagCriteria) DirectionAppliesToSource() bool {
	if c.Direction == "" {
		return true
	}
	upper := strings.ToUpper(c.Direction)
	return upper == "SRC" || upper == "EITHER"
}

// DirectionAppliesToDestination returns whether this tag applies to destination flow
func (c *TagCriteria) DirectionAppliesToDestination() bool {
	if c.Direction == "" {
		return true
	}
	upper := strings.ToUpper(c.Direction)
	return upper == "DST" || upper == "EITHER"
}

// UpdateFromUserJSON updates the model with fields received from the user JSON
// - returns whether success, and a map of keys with errors, and user-friendly error messages
// - if there are any errors, this TagCriteria's state should be disregarded
// - this mainly tests the data types fed in - true validation should be performed by Validate()
func (c *TagCriteria) UpdateFromUserJSON(fields map[string]interface{}) (bool, map[string]string) {
	ret := make(map[string]string)

	if direction, found := fields["direction"]; found {
		switch v := direction.(type) {
		case string:
			upper := strings.ToUpper(v)
			if upper != "" && upper != "DST" && upper != "SRC" && upper != "EITHER" {
				ret["direction"] = "Must be 'src', 'dst', 'either'"
			} else {
				c.Direction = upper
			}
		default:
			ret["direction"] = "Must be a string"
		}
	}
	if portRanges, found := fields["port"]; found {
		if v := toStringArray(portRanges); v != nil {
			c.PortRanges = v
		} else {
			ret["port"] = "Must be an array of strings"
		}
	}
	if protocols, found := fields["protocol"]; found {
		if v := toUint32Array(protocols); v != nil {
			c.Protocols = v
		} else {
			ret["protocol"] = "Must be an array of non-negative integers"
		}
	}
	if asnRanges, found := fields["asn"]; found {
		if v := toStringArray(asnRanges); v != nil {
			c.ASNRanges = v
		} else {
			ret["asn"] = "Must be an array of strings"
		}
	}
	if vlanRanges, found := fields["vlans"]; found {
		if v := toStringArray(vlanRanges); v != nil {
			c.VLanRanges = v
		} else {
			ret["vlans"] = "Must be an array of strings"
		}
	}
	if lasthopASNNames, found := fields["lasthop_as_name"]; found {
		if v := toStringArray(lasthopASNNames); v != nil {
			c.LastHopASNNames = v
		} else {
			ret["lasthop_as_name"] = "Must be an array of strings"
		}
	}
	if nexthopASNRanges, found := fields["nexthop_asn"]; found {
		if v := toStringArray(nexthopASNRanges); v != nil {
			c.NextHopASNRanges = v
		} else {
			ret["nexthop_asn"] = "Must be an array of strings"
		}
	}
	if nexthopASNames, found := fields["nexthop_as_name"]; found {
		if v := toStringArray(nexthopASNames); v != nil {
			c.NextHopASNNames = v
		} else {
			ret["nexthop_as_name"] = "Must be an array of strings"
		}
	}
	if bgpASPath, found := fields["bgp_aspath"]; found {
		if v := toStringArray(bgpASPath); v != nil {
			c.BGPASPaths = v
		} else {
			ret["bgp_aspath"] = "Must be an array of strings"
		}
	}
	if bgpCommunity, found := fields["bgp_community"]; found {
		if v := toStringArray(bgpCommunity); v != nil {
			c.BGPCommunities = v
		} else {
			ret["bgp_community"] = "Must be an array of strings"
		}
	}
	if tcpFlags, found := fields["tcp_flags"]; found {
		if b64Val, err := strconv.ParseUint(fmt.Sprintf("%v", tcpFlags), 10, 32); err != nil {
			ret["tcp_flags"] = "Must be an integer between 0-255"
		} else {
			if b64Val < 0 || b64Val > 255 {
				ret["tcp_flags"] = "Must be an integer between 0-255"
			} else {
				c.TCPFlags = uint32(b64Val)
			}
		}
	}
	if ipAddresses, found := fields["addr"]; found {
		if v := toStringArray(ipAddresses); v != nil {
			c.IPAddresses = v
		} else {
			ret["addr"] = "Must be an array of strings"
		}
	}
	if macAddresses, found := fields["mac"]; found {
		if v := toStringArray(macAddresses); v != nil {
			c.MACAddresses = v
		} else {
			ret["mac"] = "Must be an array of strings"
		}
	}
	if countryCodes, found := fields["country"]; found {
		if v := toStringArray(countryCodes); v != nil {
			c.CountryCodes = v
		} else {
			ret["country"] = "Must be an array of strings"
		}
	}
	if siteNameRegexes, found := fields["site"]; found {
		if v := toStringArray(siteNameRegexes); v != nil {
			c.SiteNameRegexes = v
		} else {
			ret["site"] = "Must be an array of strings"
		}
	}
	if deviceTypeRegexes, found := fields["device_type"]; found {
		if v := toStringArray(deviceTypeRegexes); v != nil {
			c.DeviceTypeRegexes = v
		} else {
			ret["device_type"] = "Must be an array of strings"
		}
	}
	if interfaceNameRegexes, found := fields["interface_name"]; found {
		if v := toStringArray(interfaceNameRegexes); v != nil {
			c.InterfaceNameRegexes = v
		} else {
			ret["interface_name"] = "Must be an array of strings"
		}
	}
	if deviceNameRegexes, found := fields["device_name"]; found {
		if v := toStringArray(deviceNameRegexes); v != nil {
			c.DeviceNameRegexes = v
		} else {
			ret["device_name"] = "Must be an array of strings"
		}
	}
	if nextHopIPAddresses, found := fields["nexthop"]; found {
		if v := toStringArray(nextHopIPAddresses); v != nil {
			c.NextHopIPAddresses = v
		} else {
			ret["nexthop"] = "Must be an array of strings"
		}
	}

	// FYI: doesn't support flex columns

	return len(ret) == 0, ret
}

// ensure and convert the input interface{} to []string, returning nil if invalid
func toStringArray(in interface{}) []string {
	switch typedIn := in.(type) {
	case []interface{}:
		ret := make([]string, 0, len(typedIn))
		for i := range typedIn {
			switch typedVal := typedIn[i].(type) {
			case string:
				ret = append(ret, typedVal)
			default:
				return nil
			}
		}
		return ret
	default:
		return nil
	}
}

// ensure and convert the input interface{} to []uint32, returning nil if invalid
func toUint32Array(in interface{}) []uint32 {
	switch typedIn := in.(type) {
	case []interface{}:
		ret := make([]uint32, 0, len(typedIn))
		for i := range typedIn {
			b64Val, err := strconv.ParseUint(fmt.Sprintf("%v", typedIn[i]), 10, 32)
			if err != nil {
				return nil
			}
			ret = append(ret, uint32(b64Val))
		}
		return ret
	default:
		return nil
	}
}
