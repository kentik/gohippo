package hippo

import (
	"strings"
)

// Normalize normalizes the fields for comparison
func (c *FlexStringCriteria) Normalize() {
	c.Action = strings.ToLower(c.Action)
	c.Value = strings.ToLower(c.Value)
}
