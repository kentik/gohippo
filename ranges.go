package hippo

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Uint8Slice attaches the methods of sort.Interface to []uint8, sorting in increasing order.
type Uint8Slice []uint8

// Len returns length
func (s Uint8Slice) Len() int { return len(s) }

// Less returns whether value at i is less than value at j
func (s Uint8Slice) Less(i, j int) bool { return s[i] < s[j] }

// Swap swaps values at two indices
func (s Uint8Slice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Sort sorts
func (s Uint8Slice) Sort() { sort.Sort(s) }

// Int32Slice attaches the methods of sort.Interface to []int32, sorting in increasing order.
type Int32Slice []int32

// Len returns length
func (s Int32Slice) Len() int { return len(s) }

// Less returns whether value at i is less than value at j
func (s Int32Slice) Less(i, j int) bool { return s[i] < s[j] }

// Swap swaps values at two indices
func (s Int32Slice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Sort sorts
func (s Int32Slice) Sort() { sort.Sort(s) }

// ASNRangesSlice is a sortable slice of ASNRanges
type ASNRangesSlice []ASNRange

// Len returns length
func (s ASNRangesSlice) Len() int { return len(s) }

// Less returns whether value at i is less than value at j
func (s ASNRangesSlice) Less(i, j int) bool {
	if s[i].Start == s[j].Start {
		return s[i].End < s[j].End
	}
	return s[i].Start < s[j].Start
}

// Swap swaps values at two indices
func (s ASNRangesSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Sort sorts
func (s ASNRangesSlice) Sort() { sort.Sort(s) }

// String returns a formatted string
// note: the slice is sorted as a result
func (s ASNRangesSlice) String() string {
	var buf bytes.Buffer
	s.Sort()
	for i, asnRange := range s {
		if i > 0 {
			buf.WriteString(",")
		}
		if asnRange.Start == asnRange.End {
			buf.WriteString(fmt.Sprintf("%d", asnRange.Start))
		} else {
			buf.WriteString(fmt.Sprintf("%d-%d", asnRange.Start, asnRange.End))
		}
	}
	return buf.String()
}

// ToStringArray returns a sorted string array
func (s ASNRangesSlice) ToStringArray() []string {
	s.Sort()
	ret := make([]string, 0)
	for _, asnRange := range s {
		if asnRange.Start == asnRange.End {
			ret = append(ret, fmt.Sprintf("%d", asnRange.Start))
		} else {
			ret = append(ret, fmt.Sprintf("%d-%d", asnRange.Start, asnRange.End))
		}
	}
	return ret
}

// PortRangesSlice is a sortable PortRange slice
type PortRangesSlice []PortRange

// Len returns length
func (s PortRangesSlice) Len() int { return len(s) }

// Less returns whether value at i is less than value at j
func (s PortRangesSlice) Less(i, j int) bool {
	if s[i].Start == s[j].Start {
		return s[i].End < s[j].End
	}
	return s[i].Start < s[j].Start
}

// Swap swaps values at two indices
func (s PortRangesSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Sort sorts
func (s PortRangesSlice) Sort() { sort.Sort(s) }

// String returns a formatted string
// note: the slice is sorted as a result
func (s PortRangesSlice) String() string {
	var buf bytes.Buffer
	s.Sort()
	for i, asnRange := range s {
		if i > 0 {
			buf.WriteString(",")
		}
		if asnRange.Start == asnRange.End {
			buf.WriteString(fmt.Sprintf("%d", asnRange.Start))
		} else {
			buf.WriteString(fmt.Sprintf("%d-%d", asnRange.Start, asnRange.End))
		}
	}
	return buf.String()
}

// ToStringArray returns a sorted string array
func (s PortRangesSlice) ToStringArray() []string {
	s.Sort()
	ret := make([]string, 0)
	for _, asnRange := range s {
		if asnRange.Start == asnRange.End {
			ret = append(ret, fmt.Sprintf("%d", asnRange.Start))
		} else {
			ret = append(ret, fmt.Sprintf("%d-%d", asnRange.Start, asnRange.End))
		}
	}
	return ret
}

// VLanRangesSlice is a sortable slice of VLanRanges
type VLanRangesSlice []VLanRange

// Len returns length
func (s VLanRangesSlice) Len() int { return len(s) }

// Less returns whether value at i is less than value at j
func (s VLanRangesSlice) Less(i, j int) bool {
	if s[i].Start == s[j].Start {
		return s[i].End < s[j].End
	}
	return s[i].Start < s[j].Start
}

// Swap swaps values at two indices
func (s VLanRangesSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Sort sorts
func (s VLanRangesSlice) Sort() { sort.Sort(s) }

// String returns a formatted string
// note: the slice is sorted as a result
func (s VLanRangesSlice) String() string {
	var buf bytes.Buffer
	s.Sort()
	for i, vlanRange := range s {
		if i > 0 {
			buf.WriteString(",")
		}
		if vlanRange.Start == vlanRange.End {
			buf.WriteString(fmt.Sprintf("%d", vlanRange.Start))
		} else {
			buf.WriteString(fmt.Sprintf("%d-%d", vlanRange.Start, vlanRange.End))
		}
	}
	return buf.String()
}

// ToStringArray returns a sorted string array
func (s VLanRangesSlice) ToStringArray() []string {
	s.Sort()
	ret := make([]string, 0)
	for _, vlanRange := range s {
		if vlanRange.Start == vlanRange.End {
			ret = append(ret, fmt.Sprintf("%d", vlanRange.Start))
		} else {
			ret = append(ret, fmt.Sprintf("%d-%d", vlanRange.Start, vlanRange.End))
		}
	}
	return ret
}

// ParseASNs returns the parsed ASN list, along with an error message to show to customer
// - error values are skipped, so you can ignore the error if you like
func ParseASNs(asns []string, errMsg *string) []ASNRange {
	ret := make([]ASNRange, 0, 0)

	for _, str := range asns {
		parts := strings.Split(str, "-")
		if len(parts) == 1 {
			// regular ASN
			if val, ok := parseUint32(str); ok {
				ret = append(ret, ASNRange{Start: val, End: val})
			} else {
				appendErrorMessage(errMsg, "Invalid ASN: '%s'", str)
			}
		} else if len(parts) == 2 {
			// ASN range
			start, startOk := parseUint32(parts[0])
			end, endOk := parseUint32(parts[1])
			if startOk && endOk && end >= start {
				ret = append(ret, ASNRange{Start: start, End: end})
			} else {
				if !startOk {
					appendErrorMessage(errMsg, "Invalid ASN: '%s'", parts[0])
				}
				if !endOk {
					appendErrorMessage(errMsg, "Invalid ASN: '%s'", parts[1])
				}
				if startOk && endOk {
					// valid ASNs, must not be a valid range
					appendErrorMessage(errMsg, "Invalid ASN range: '%s'", str)
				}
			}
		} else {
			// bad format - skip
			appendErrorMessage(errMsg, "Invalid ASN range: '%s'", str)
			continue
		}
	}
	return ret
}

// ParseVLans parses a string array of VLan ranges into []VLanRange,
// skipping invalid entries, and returning an error string meant
// for the user, but can be ignored internally
func ParseVLans(vlans []string, errMsg *string) []VLanRange {
	ret := make([]VLanRange, 0)

	for _, str := range vlans {
		parts := strings.Split(str, "-")
		if len(parts) == 1 {
			// single VLAN
			if val, ok := parseUint32(str); ok {
				if val >= 0 && val <= 4095 {
					ret = append(ret, VLanRange{Start: val, End: val})
				}
			} else {
				appendErrorMessage(errMsg, "invalid VLAN: %s", str)
			}
		} else if len(parts) == 2 {
			// range
			start, startOk := parseUint32(parts[0])
			end, endOk := parseUint32(parts[1])
			if startOk && endOk && end >= start && start >= 0 && start <= 4095 && end >= 0 && end <= 4095 {
				// valid range
				ret = append(ret, VLanRange{Start: start, End: end})
			} else {
				if !startOk || start < 0 || start > 4095 {
					appendErrorMessage(errMsg, "invalid VLAN: %d", start)
				}
				if !endOk || end < 0 || end > 4095 {
					appendErrorMessage(errMsg, "invalid VLAN: %d", end)
				}
				if start > end {
					appendErrorMessage(errMsg, "invalid VLAN range: start (%d) is greater than end (%d)", start, end)
				}
			}
		} else {
			// bad format - skip
			appendErrorMessage(errMsg, "invalid VLAN: %s", str)
			continue
		}
	}
	return ret
}

// ParsePorts parses the input string array of port ranges,
// returning a PortRangesSlice with invalid entries removed,
// as well as an error string meant for the user, and can be
// ignored internally
func ParsePorts(ports []string, errMsg *string) []PortRange {
	ret := make([]PortRange, 0, 0)

	for _, str := range ports {
		parts := strings.Split(str, "-")
		if len(parts) == 1 {
			if val, ok := parseUint32(str); ok {
				ret = append(ret, PortRange{Start: val, End: val})
			} else {
				appendErrorMessage(errMsg, "invalid port: %s", str)
			}
		} else if len(parts) == 2 {
			start, startOk := parseUint32(parts[0])
			end, endOk := parseUint32(parts[1])
			if startOk && endOk && end >= start {
				ret = append(ret, PortRange{Start: start, End: end})
			} else {
				// error
				if !startOk {
					appendErrorMessage(errMsg, "invalid port: %s", parts[0])
				}
				if !endOk {
					appendErrorMessage(errMsg, "invalid port: %s", parts[1])
				}
				if startOk && endOk && start > end {
					appendErrorMessage(errMsg, "invalid port range: start (%d) is greater than end (%d)", start, end)
				}
			}
		} else {
			// bad format - skip
			appendErrorMessage(errMsg, "invalid port range: '%s'", str)
			continue
		}
	}
	return ret
}

// ParseProtocols validates the input int32 array, making sure each is 0-255
func ParseProtocols(protocols []uint32, errMsg *string) []uint32 {
	ret := make([]uint32, 0, len(protocols))
	for _, protocol := range protocols {
		if protocol < 0 || protocol > 255 {
			// invalid - skip it
			appendErrorMessage(errMsg, "invalid protocol: %d is not between 0-255", protocol)
		} else {
			// valid
			ret = append(ret, protocol)
		}
	}
	return ret
}

// return parsed uint32 and whether it was successful
func parseUint32(str string) (uint32, bool) {
	i32, err := strconv.ParseUint(strings.TrimSpace(str), 10, 32)
	if err == nil { // ignore error
		return uint32(i32), true
	}
	return 0, false
}

// return parsed uint64 and whether it was successful
func parseUint64(str string) (uint64, bool) {
	i64, err := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
	if err == nil {
		return i64, true
	}
	return 0, false
}

func appendErrorMessage(message *string, newError string, fmtStrs ...interface{}) {
	if message == nil {
		return
	}
	if len(*message) > 0 {
		*message = *message + "; " + fmt.Sprintf(newError, fmtStrs...)
	} else {
		*message = fmt.Sprintf(newError, fmtStrs...)
	}
}
