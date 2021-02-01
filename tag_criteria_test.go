package hippo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateFromUserJSON(t *testing.T) {
	// sut1 is updated via map, compared to manually updated sut2

	sut1 := TagCriteria{
		Direction:            "SRC",
		PortRanges:           []string{"1-5", "2-4"},
		Protocols:            []uint32{uint32(1), uint32(2)},
		ASNRanges:            []string{"1-10", "2-5"},
		LastHopASNNames:      []string{"a", "b"},
		NextHopASNRanges:     []string{"1-11", "2-22"},
		NextHopASNNames:      []string{"a", "b"},
		BGPASPaths:           []string{"a", "b"},
		BGPCommunities:       []string{"a", "b"},
		TCPFlags:             uint32(5),
		IPAddresses:          []string{"a", "b"},
		MACAddresses:         []string{"a", "b"},
		CountryCodes:         []string{"a", "b"},
		SiteNameRegexes:      []string{"a", "b"},
		DeviceTypeRegexes:    []string{"a", "b"},
		InterfaceNameRegexes: []string{"a", "b"},
		DeviceNameRegexes:    []string{"a", "b"},
		NextHopIPAddresses:   []string{"a", "b"},
		VLanRanges:           []string{"1-100", "33"},
	}

	sut2 := TagCriteria{
		Direction:            "SRC",
		PortRanges:           []string{"1-5", "2-4"},
		Protocols:            []uint32{uint32(1), uint32(2)},
		ASNRanges:            []string{"1-10", "2-5"},
		LastHopASNNames:      []string{"a", "b"},
		NextHopASNRanges:     []string{"1-11", "2-22"},
		NextHopASNNames:      []string{"a", "b"},
		BGPASPaths:           []string{"a", "b"},
		BGPCommunities:       []string{"a", "b"},
		TCPFlags:             uint32(5),
		IPAddresses:          []string{"a", "b"},
		MACAddresses:         []string{"a", "b"},
		CountryCodes:         []string{"a", "b"},
		SiteNameRegexes:      []string{"a", "b"},
		DeviceTypeRegexes:    []string{"a", "b"},
		InterfaceNameRegexes: []string{"a", "b"},
		DeviceNameRegexes:    []string{"a", "b"},
		NextHopIPAddresses:   []string{"a", "b"},
		VLanRanges:           []string{"1-100", "33"},
	}

	// no changes
	assert.Equal(t, sut1.GenerateHash(), sut2.GenerateHash())

	runTest := func(updates map[string]interface{}) {
		valid, errs := sut1.UpdateFromUserJSON(updates)
		assert.True(t, valid)
		assert.Equal(t, 0, len(errs))
		assert.Equal(t, sut2.GenerateHash(), sut1.GenerateHash())
	}

	// set direction via update
	sut2.Direction = "DST"
	runTest(map[string]interface{}{
		"direction": "DST",
	})

	sut2.PortRanges = []string{"1-2", "3-4"}
	runTest(map[string]interface{}{
		"port": []interface{}{"3-4", "1-2"},
	})

	sut2.Protocols = []uint32{5, 10}
	runTest(map[string]interface{}{"protocol": []interface{}{10, 5}})

	sut2.ASNRanges = []string{"123-456", "500-1000"}
	runTest(map[string]interface{}{"asn": []interface{}{"500-1000", "123-456"}})

	sut2.VLanRanges = []string{"1-10", "20-25"}
	runTest(map[string]interface{}{"vlans": []interface{}{"20-25", "1-10"}})

	sut2.LastHopASNNames = []string{"FOO", "BAR"}
	runTest(map[string]interface{}{"lasthop_as_name": []interface{}{"BAR", "FOO"}})

	sut2.NextHopASNRanges = []string{"8-9", "2-5"}
	runTest(map[string]interface{}{"nexthop_asn": []interface{}{"2-5", "8-9"}})

	sut2.NextHopASNNames = []string{"B", "A"}
	runTest(map[string]interface{}{"nexthop_as_name": []interface{}{"A", "B"}})

	sut2.BGPASPaths = []string{"D", "C"}
	runTest(map[string]interface{}{"bgp_aspath": []interface{}{"C", "D"}})

	sut2.BGPCommunities = []string{"BBB", "AAA"}
	runTest(map[string]interface{}{"bgp_community": []interface{}{"AAA", "BBB"}})

	sut2.TCPFlags = 222
	runTest(map[string]interface{}{"tcp_flags": interface{}(222)})

	sut2.IPAddresses = []string{"2.3.4.5/32", "1.2.3.4/31"}
	runTest(map[string]interface{}{"addr": []interface{}{interface{}("1.2.3.4/31"), interface{}("2.3.4.5/32")}})

	sut2.MACAddresses = []string{"FOO:BAR", "HEY:NOW"}
	runTest(map[string]interface{}{"mac": []interface{}{interface{}("HEY:NOW"), interface{}("FOO:BAR")}})

	sut2.CountryCodes = []string{"US", "EN"}
	runTest(map[string]interface{}{"country": []interface{}{interface{}("EN"), interface{}("US")}})

	sut2.SiteNameRegexes = []string{"IIII", "JJJJ"}
	runTest(map[string]interface{}{"site": []interface{}{interface{}("JJJJ"), interface{}("IIII")}})

	sut2.DeviceTypeRegexes = []string{"KKKK", "LLLL"}
	runTest(map[string]interface{}{"device_type": []interface{}{interface{}("LLLL"), interface{}("KKKK")}})

	sut2.InterfaceNameRegexes = []string{"OOOO", "PPPP"}
	runTest(map[string]interface{}{"interface_name": []interface{}{interface{}("PPPP"), interface{}("OOOO")}})

	sut2.DeviceNameRegexes = []string{"UUUU", "VVVV"}
	runTest(map[string]interface{}{"device_name": []interface{}{interface{}("VVVV"), interface{}("UUUU")}})

	sut2.NextHopIPAddresses = []string{"8.8.8.8", "9.9.9.9"}
	runTest(map[string]interface{}{"nexthop": []interface{}{interface{}("9.9.9.9"), interface{}("8.8.8.8")}})
}

func TestEnsureAndSortFlexStringMatchArray(t *testing.T) {
	array := []FlexStringCriteria{
		FlexStringCriteria{
			Action: "exact",
			Value:  "foobar",
		},
		FlexStringCriteria{
			Action: "prefIX",
			Value:  "http",
		},
		FlexStringCriteria{
			Action: "EXACT",
			Value:  "ABCdef",
		},
		FlexStringCriteria{
			Action: "prefix",
			Value:  "eieio",
		},
	}

	ensureAndSortFlexStringMatchArray(&array)

	assert.Equal(t, 4, len(array))
	assert.Equal(t, "exact", array[0].Action)
	assert.Equal(t, "abcdef", array[0].Value)
	assert.Equal(t, "exact", array[1].Action)
	assert.Equal(t, "foobar", array[1].Value)
	assert.Equal(t, "prefix", array[2].Action)
	assert.Equal(t, "eieio", array[2].Value)
	assert.Equal(t, "prefix", array[3].Action)
	assert.Equal(t, "http", array[3].Value)

	// try with nil
	array = nil
	ensureAndSortFlexStringMatchArray(&array)
	assert.NotNil(t, array)
	assert.Equal(t, 0, len(array))
}
