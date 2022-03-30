package hippo

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/kentik/patricia"
)

// regexp for validating BGP community and ASPath
var _bgpStringRegex *regexp.Regexp

// initialize the BGP strings regex
// - this is called by init()
func initValidationRegexes() {
	var err error
	_bgpStringRegex, err = regexp.Compile("^[" + regexp.QuoteMeta("[]*:_^$|.0123456789()+?, -") + "]+$")
	if err != nil {
		panic("Error compiling BGP validation regexp")
	}
}

// Validate validates the criteria, returning true if valid, and a customer-friendly error string if false
// - top-level message stored in "" key
func (c *TagCriteria) Validate(isPopulator bool) (bool, map[string]string) {
	ret := make(map[string]string)
	var errMsg string
	hasCriteria := false

	// Direction = 1
	direction := strings.ToUpper(c.Direction)
	if isPopulator {
		// populator
		if direction != "" && direction != "SRC" && direction != "DST" && direction != "EITHER" {
			ret["direction"] = "Must be 'src', 'dst', or 'either'"
		}
	} else {
		// tag
		if direction != "" && direction != "EITHER" {
			ret["direction"] = "Must be '' or 'either' for tags"
		}
	}

	// PortRanges = 2
	if len(c.PortRanges) > 0 {
		hasCriteria = true
		if len(c.PortRanges) > 100 {
			ret["port"] = fmt.Sprintf("Too many port ranges: found %d, allowed: 100", len(c.PortRanges))
		} else {
			if ParsePorts(c.PortRanges, &errMsg); errMsg != "" {
				ret["port"] = errMsg
			}
		}
	}

	// Protocols = 3
	if len(c.Protocols) > 0 {
		hasCriteria = true
		if len(c.Protocols) > 100 {
			ret["protocol"] = fmt.Sprintf("Too many protocols: found %d, allowed: 100", len(c.Protocols))
		} else {
			if ParseProtocols(c.Protocols, &errMsg); errMsg != "" {
				ret["protocol"] = errMsg
			}
		}
	}

	// ASNRanges = 4
	if len(c.ASNRanges) > 0 {
		hasCriteria = true
		if len(c.ASNRanges) > 100 {
			ret["asn"] = fmt.Sprintf("Too many ASN ranges: found %d, allowed: 100", len(c.ASNRanges))
		} else {
			if ParseASNs(c.ASNRanges, &errMsg); errMsg != "" {
				ret["asn"] = errMsg
			}
		}
	}

	// VLanRanges = 5
	if len(c.VLanRanges) > 0 {
		hasCriteria = true
		// parse the vlans and make sure we have the same number
		if ParseVLans(c.VLanRanges, &errMsg); errMsg != "" {
			ret["vlans"] = errMsg
		}
	}

	// LastHopASNNames = 6
	if len(c.LastHopASNNames) > 0 {
		hasCriteria = true
		if len(c.LastHopASNNames) > 500 {
			ret["lasthop_as_name"] = fmt.Sprintf("Too many last-hop AS names: found %d, allowed: 500", len(c.LastHopASNNames))
		}
	}

	// NextHopASNRanges = 7
	if len(c.NextHopASNRanges) > 0 {
		hasCriteria = true
		if len(c.NextHopASNRanges) > 100 {
			ret["nexthop_asn"] = fmt.Sprintf("Too many next-hop ASN ranges: found %d, allowed: 100", len(c.NextHopASNRanges))
		} else {
			if ParseASNs(c.NextHopASNRanges, &errMsg); errMsg != "" {
				ret["nexthop_asn"] = errMsg
			}
		}
	}

	// NextHopASNNames = 8
	if len(c.NextHopASNNames) > 0 {
		hasCriteria = true
		if len(c.NextHopASNNames) > 500 {
			ret["nexthop_as_name"] = fmt.Sprintf("Too many next-hop AS names: found %d, allowed: 500", len(c.NextHopASNNames))
		}
	}

	// BGPASPaths = 9
	if len(c.BGPASPaths) > 0 {
		hasCriteria = true
		if SanitizeBGPASPaths(c.BGPASPaths, &errMsg); errMsg != "" {
			ret["bgp_aspath"] = errMsg
		}
	}

	// BGPCommunities = 10
	if len(c.BGPCommunities) > 0 {
		hasCriteria = true
		if SanitizeBGPCommunities(c.BGPCommunities, &errMsg); errMsg != "" {
			ret["bgp_community"] = errMsg
		}
	}

	// TCPFlags = 11
	if c.TCPFlags != 0 {
		hasCriteria = true
		if SanitizeTCPFlags(c.TCPFlags, &errMsg); errMsg != "" {
			ret["tcp_flags"] = errMsg
		}
	}

	// IPAddresses = 12
	if len(c.IPAddresses) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.IPAddresses {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["addr"] = "Invalid IP address(es)"
			}
		}
	}

	// MACAddresses = 13
	if len(c.MACAddresses) > 0 {
		hasCriteria = true
		if SanitizeMACAddresses(c.MACAddresses, &errMsg); errMsg != "" {
			ret["mac"] = errMsg
		}
	}

	// CountryCodes = 14
	if len(c.CountryCodes) > 0 {
		hasCriteria = true
		// someday, load up the country codes, for now - don't worry about it
	}

	// SiteNameRegexes = 15
	if len(c.SiteNameRegexes) > 0 {
		hasCriteria = true
		if len(c.SiteNameRegexes) > 500 {
			ret["site"] = fmt.Sprintf("Too many site names: found %d, allowed: 500", len(c.SiteNameRegexes))
		}
	}

	// DeviceTypeRegexes = 16
	if len(c.DeviceTypeRegexes) > 0 {
		hasCriteria = true
		if len(c.DeviceTypeRegexes) > 100 {
			ret["device_type"] = fmt.Sprintf("Too many device types: found %d, allowed: 100", len(c.DeviceTypeRegexes))
		}
	}

	// InterfaceNameRegexes = 17
	if len(c.InterfaceNameRegexes) > 0 {
		hasCriteria = true
	}

	// DeviceNameRegexes = 18
	if len(c.DeviceNameRegexes) > 0 {
		hasCriteria = true
	}

	// NextHopIPAddresses = 19
	if len(c.NextHopIPAddresses) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.NextHopIPAddresses {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["nexthop"] = "Invalid next-hop IP address(es)"
				break
			}
		}
	}

	if len(c.Str00) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str00); !success {
			ret["str00"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str01) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str01); !success {
			ret["str01"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str02) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str02); !success {
			ret["str02"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str03) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str03); !success {
			ret["str03"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str04) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str04); !success {
			ret["str04"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str05) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str05); !success {
			ret["str05"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str06) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str06); !success {
			ret["str06"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str07) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str07); !success {
			ret["str07"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str08) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str08); !success {
			ret["str08"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str09) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str09); !success {
			ret["str09"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str10) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str10); !success {
			ret["str10"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str11) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str11); !success {
			ret["str11"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str12) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str12); !success {
			ret["str12"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str13) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str13); !success {
			ret["str13"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str14) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str14); !success {
			ret["str14"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str15) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str15); !success {
			ret["str15"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str16) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str16); !success {
			ret["str16"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str17) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str17); !success {
			ret["str17"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str18) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str18); !success {
			ret["str18"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str19) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str19); !success {
			ret["str19"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str20) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str20); !success {
			ret["str20"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str21) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str21); !success {
			ret["str21"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str22) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str22); !success {
			ret["str22"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str23) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str23); !success {
			ret["str23"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str24) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str24); !success {
			ret["str24"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str25) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str25); !success {
			ret["str25"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str26) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str26); !success {
			ret["str26"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str27) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str27); !success {
			ret["str27"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str28) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str28); !success {
			ret["str28"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str29) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str29); !success {
			ret["str29"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str30) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str30); !success {
			ret["str30"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str31) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str31); !success {
			ret["str31"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Str32) > 0 {
		if success, errMsg := validateFlexStringCriteria(c.Str32); !success {
			ret["str32"] = errMsg
		} else {
			hasCriteria = true
		}
	}

	if len(c.Int00) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int00); errMsg != "" {
			ret["int00"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int01) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int01); errMsg != "" {
			ret["int01"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int02) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int02); errMsg != "" {
			ret["int02"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int03) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int03); errMsg != "" {
			ret["int03"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int04) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int04); errMsg != "" {
			ret["int04"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int05) > 0 {
		if _, errMsg := NewFlexUint32RangesFromStrings(c.Int05); errMsg != "" {
			ret["int05"] = errMsg
		} else {
			hasCriteria = true
		}
	}

	if len(c.AppProtocol) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.AppProtocol); errMsg != "" {
			ret["app_protocol"] = errMsg
		} else {
			hasCriteria = true
		}
	}

	if len(c.Int6400) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.Int6400); errMsg != "" {
			ret["int64_00"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int6401) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.Int6401); errMsg != "" {
			ret["int64_01"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int6402) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.Int6402); errMsg != "" {
			ret["int64_02"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int6403) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.Int6403); errMsg != "" {
			ret["int64_03"] = errMsg
		} else {
			hasCriteria = true
		}
	}
	if len(c.Int6404) > 0 {
		if _, errMsg := NewFlexUint64RangesFromStrings(c.Int6404); errMsg != "" {
			ret["int64_04"] = errMsg
		} else {
			hasCriteria = true
		}
	}

	if len(c.Inet00) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.Inet00 {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["inet_00"] = "Invalid IP address(es)"
			}
		}
	}
	if len(c.Inet01) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.Inet01 {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["inet_01"] = "Invalid IP address(es)"
			}
		}
	}
	if len(c.Inet02) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.Inet02 {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["inet_02"] = "Invalid IP address(es)"
			}
		}
	}
	if len(c.Inet03) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.Inet03 {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["inet_03"] = "Invalid IP address(es)"
			}
		}
	}
	if len(c.Inet04) > 0 {
		hasCriteria = true
		for _, ipAddress := range c.Inet04 {
			if _, _, err := patricia.ParseIPFromString(ipAddress); err != nil {
				ret["inet_04"] = "Invalid IP address(es)"
			}
		}
	}

	if !hasCriteria {
		ret[""] = "Missing criteria"
	}
	return len(ret) == 0, ret
}

// validate flex string criteria, returning success, and error message
func validateFlexStringCriteria(flexCriteria []FlexStringCriteria) (bool, string) {
	errMsg := ""
	for _, crit := range flexCriteria {
		if !strings.EqualFold(crit.Action, "exact") && !strings.EqualFold(crit.Action, "prefix") {
			appendErrorMessage(&errMsg, "Invalid 'action'")
		}
		if crit.Value == "" {
			appendErrorMessage(&errMsg, "missing 'value'")
		} else if len(crit.Value) > 200 {
			appendErrorMessage(&errMsg, "'value' too long")
		}
	}
	return errMsg == "", errMsg
}

// SanitizeBGPASPaths sanitizes a list of BGPASPaths, returning a filtered list and an error message
func SanitizeBGPASPaths(bgpASPaths []string, errMsg *string) []string {
	ret := make([]string, 0, len(bgpASPaths))

	for _, asPath := range bgpASPaths {
		if _bgpStringRegex.MatchString(asPath) {
			ret = append(ret, asPath)
		} else {
			appendErrorMessage(errMsg, "invalid BGP AS Path: '%s'", asPath)
		}
	}
	return ret
}

// SanitizeBGPCommunities sanitizes a list of BGPCommunities, returning a filtered list and an error message
func SanitizeBGPCommunities(bgpCommunities []string, errMsg *string) []string {
	ret := make([]string, 0, len(bgpCommunities))

	for _, community := range bgpCommunities {
		if _bgpStringRegex.MatchString(community) {
			ret = append(ret, community)
		} else {
			appendErrorMessage(errMsg, "invalid BGP community: '%s'", community)
		}
	}
	return ret
}

// SanitizeTCPFlags sanitizes TCP flags, setting it to 0 if invalid, and returning an error message
func SanitizeTCPFlags(tcpFlags uint32, errMsg *string) uint32 {
	if tcpFlags > 255 {
		*errMsg = "invalid tcp flags: must be 0-255"
		return 0
	}
	return tcpFlags
}

// SanitizeMACAddresses sanitizes mac addresses, removing invalid ones, and updating the error message
func SanitizeMACAddresses(macAddresses []string, errMsg *string) []string {
	ret := make([]string, 0, len(macAddresses))
	for _, mac := range macAddresses {
		if _, err := net.ParseMAC(mac); err == nil {
			ret = append(ret, mac)
		} else {
			appendErrorMessage(errMsg, "invalid MAC: %s", mac)
		}
	}
	return ret
}
