package hippo

import (
	fmt "fmt"
	"strings"
)

// NewFlexUint64RangesFromStrings parses the input string into a Uint64Range
// - returns error message suitable for user
func NewFlexUint64RangesFromStrings(rangeStrs []string) ([]FlexUint64Range, string) {
	ret := make([]FlexUint64Range, 0, len(rangeStrs))
	for _, rangeStr := range rangeStrs {
		parts := strings.Split(rangeStr, "-")
		if len(parts) == 1 {
			parts[0] = strings.TrimSpace(parts[0])
			parsedVal, ok := parseUint64(parts[0])
			if !ok {
				return nil, "Invalid input"
			}
			ret = append(ret, FlexUint64Range{Start: parsedVal, End: parsedVal})
		} else if len(parts) == 2 {
			parts[0] = strings.TrimSpace(parts[0])
			parts[1] = strings.TrimSpace(parts[1])

			parsedVal, ok := parseUint64(parts[0])
			if !ok {
				return nil, "Invalid start"
			}
			newRange := FlexUint64Range{Start: parsedVal}

			parsedVal, ok = parseUint64(parts[1])
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
