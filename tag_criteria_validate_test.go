package hippo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeBGPCommunities(t *testing.T) {
	bgpCommunities := []string{"foo", "214"}
	errMessage := ""

	sanitizedCommunities := SanitizeBGPCommunities(bgpCommunities, &errMessage)
	if !assert.Equal(t, 1, len(sanitizedCommunities)) {
		return
	}
	assert.Equal(t, "214", sanitizedCommunities[0])
	assert.Equal(t, "invalid BGP community: 'foo'", errMessage)
}

func TestSanitizeBGPASPaths(t *testing.T) {
	bgpASPaths := []string{"foo", "[[.^$.03214"}
	errMessage := ""

	sanitizedASPaths := SanitizeBGPASPaths(bgpASPaths, &errMessage)
	if !assert.Equal(t, 1, len(sanitizedASPaths)) {
		return
	}
	assert.Equal(t, "[[.^$.03214", sanitizedASPaths[0])
	assert.Equal(t, "invalid BGP AS Path: 'foo'", errMessage)
}

func TestSanitizeBGPASPaths2(t *testing.T) {
	bgpASPaths := []string{"_(19|62|81|188|210|264|549|555|784|792|794|803|1280|1313)_"}
	errMessage := ""

	sanitizedASPaths := SanitizeBGPASPaths(bgpASPaths, &errMessage)
	if !assert.Equal(t, 1, len(sanitizedASPaths)) {
		return
	}
	assert.Equal(t, "_(19|62|81|188|210|264|549|555|784|792|794|803|1280|1313)_", sanitizedASPaths[0])
	assert.Equal(t, "", errMessage)
}

func Test_validateFlexStringCriteria(t *testing.T) {
	type args struct {
		flexCriteria []FlexStringCriteria
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"simple_valid",
			args{flexCriteria: []FlexStringCriteria{
				{Action: FlexStringActionExact, Value: "exact match"},
				{Action: FlexStringActionPrefix, Value: "prefix match"},
			}},
			true,
		},
		{
			"invalid_action",
			args{flexCriteria: []FlexStringCriteria{
				{Action: "invalid_action", Value: "bad criteria"},
			}},
			false,
		},
		{
			"mixed_with_no_value",
			args{flexCriteria: []FlexStringCriteria{
				{Action: FlexStringActionExact, Value: "valid value"},
				{Action: FlexStringActionPrefix, Value: ""},
			}},
			false,
		},
		{
			"too_long",
			args{flexCriteria: []FlexStringCriteria{
				{
					Action: FlexStringActionExact,
					Value: `this is a value that is going to be way too long, in fact it will be over 200 characters ` +
						`long which is beyond the maximum supported length of a value. Enjoy reading this really ` +
						`long string as a test case to ensure validation works as expected`,
				},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, errorString := validateFlexStringCriteria(tt.args.flexCriteria)
			if isValid != tt.want {
				t.Errorf("validateFlexStringCriteria() isValid = %v, want %v", isValid, tt.want)
			}

			if tt.want {
				assert.Equal(t, 0, len(errorString), "expected an empty error")
			} else {
				assert.NotEqual(t, 0, len(errorString), "expected non empty error")
			}
		})
	}
}

func TestTagCriteria_ValidateStrings(t *testing.T) {
	invalidCriteria := []FlexStringCriteria{
		{Action: "invalid_action", Value: "bad criteria"},
	}
	c := TagCriteria{
		Str00: invalidCriteria,
		Str01: invalidCriteria,
		Str02: invalidCriteria,
		Str03: invalidCriteria,
		Str04: invalidCriteria,
		Str05: invalidCriteria,
		Str06: invalidCriteria,
		Str07: invalidCriteria,
		Str08: invalidCriteria,
		Str09: invalidCriteria,
		Str10: invalidCriteria,
		Str11: invalidCriteria,
		Str12: invalidCriteria,
		Str13: invalidCriteria,
		Str14: invalidCriteria,
		Str15: invalidCriteria,
		Str16: invalidCriteria,
		Str17: invalidCriteria,
		Str18: invalidCriteria,
		Str19: invalidCriteria,
		Str20: invalidCriteria,
		Str21: invalidCriteria,
		Str22: invalidCriteria,
		Str23: invalidCriteria,
		Str24: invalidCriteria,
		Str25: invalidCriteria,
		Str26: invalidCriteria,
		Str27: invalidCriteria,
		Str28: invalidCriteria,
		Str29: invalidCriteria,
		Str30: invalidCriteria,
		Str31: invalidCriteria,
		Str32: invalidCriteria,
	}

	isValid, errors := c.Validate(true)
	assert.False(t, isValid)

	for i := 0; i <= 32; i++ {
		field := fmt.Sprintf("str%02d", i)
		_, errorFound := errors[field]
		assert.True(t, errorFound, "missing expected error for %s", field)
	}
}
