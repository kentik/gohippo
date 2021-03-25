package hippo

import (
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
