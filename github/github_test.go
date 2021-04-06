package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getGitHubAPIURL(t *testing.T) {
	// test that returns default GITHUB_API_URL
	s, err := getGitHubAPIURL()
	assert.NoError(t, err)
	assert.Equal(t, GITHUB_API_URL, s)

	// test that returns GITHUB_API_URL environment variable
	os.Setenv("GITHUB_API_URL", "https://api.example.com")
	s, err = getGitHubAPIURL()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com", s)

	// test that remove a single-slash path string
	os.Setenv("GITHUB_API_URL", "https://api.example.com/")
	s, err = getGitHubAPIURL()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com", s)

	// test that returns invalid url error
	os.Setenv("GITHUB_API_URL", "/")
	s, err = getGitHubAPIURL()
	assert.Empty(t, s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid url")

	// test that returns error
	for _, v := range []string{
		"https://api.example.com/foo",
		"https://api.example.com?query=value",
		"https://api.example.com#fragment",
	} {
		os.Setenv("GITHUB_API_URL", v)
		s, err = getGitHubAPIURL()
		assert.Empty(t, s)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid url")
	}
}

func Test_resolveEndpoint(t *testing.T) {
	// test that returns resolved endpoint
	s, err := resolveEndpoint("foo//bar/.///././../baz/../qux?q=a&x=y")
	assert.NoError(t, err)
	assert.Equal(t, "/foo/bar/qux?q=a&x=y", s)

	// test that returns error
	for _, v := range []string{
		"https://",
		"https://api.example.com/foo",
		"foo#fragment",
	} {
		s, err = resolveEndpoint(v)
		assert.Empty(t, s)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid endpoint")
	}
}
