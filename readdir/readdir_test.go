package readdir

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	// test that clean dirname
	r := New("/foo/../bar/", "pattern", nil)
	assert.Equal(t, r.dirname, "/bar")
}

func Test_Reader_String(t *testing.T) {
	r := New("/foo/../bar/", "pattern", nil)
	// test that return "/bar/pattern"
	assert.Equal(t, r.String(), "/bar/pattern")
}

func Test_Reader_MatchString(t *testing.T) {
	// test that compare in plain text
	r := New("/foo/../bar/", `.*.tar.gz`, nil)
	assert.True(t, r.MatchString(".*.tar.gz"))
	assert.False(t, r.MatchString("test-match.tar.gz"))

	// test that matches the specified plain text
	r = New("/foo/../bar/", `.*.tar.gz`, regexp.MustCompile(`.*.tar.gz`))
	assert.True(t, r.MatchString(".*.tar.gz"))
	assert.True(t, r.MatchString("test-match.tar.gz"))
}
