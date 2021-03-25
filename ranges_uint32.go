package hippo

import (
	"fmt"
	"strings"
)

// NewFlexUint32RangesFromStrings parses the input strings into a slice of FlexUint32Range
// - returns error message suitable for user
func NewFlexUint32RangesFromStrings(rangeStrs []string) ([]FlexUint32Range, string) {
	ret := make([]FlexUint32Range, 0, len(rangeStrs))
	for _, rangeStr := range rangeStrs {
		parts := strings.Split(rangeStr, "-")
		if len(parts) == 1 {
			parts[0] = strings.TrimSpace(parts[0])
			parsedVal, ok := parseUint32(parts[0])
			if !ok {
				return nil, "Invalid input"
			}
			ret = append(ret, FlexUint32Range{Start: parsedVal, End: parsedVal})
		} else if len(parts) == 2 {
			parts[0] = strings.TrimSpace(parts[0])
			parts[1] = strings.TrimSpace(parts[1])

			parsedVal, ok := parseUint32(parts[0])
			if !ok {
				return nil, "Invalid start"
			}
			newRange := FlexUint32Range{}
			newRange.Start = parsedVal

			parsedVal, ok = parseUint32(parts[1])
			if !ok {
				return nil, "Invalid end"
			}
			newRange.End = parsedVal

			if newRange.Start > newRange.End {
				return nil, "Start is greater than End"
			}
			ret = append(ret, newRange)
		} else {
			return nil, fmt.Sprintf("Too many parts to range string: %s", rangeStr)
		}
	}
	return ret, ""
}
