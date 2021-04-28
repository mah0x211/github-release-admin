package readdir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	// test that clean dirname
	r, err := New("/foo/../bar/", "pattern", AsPlain)
	assert.NoError(t, err)
	assert.Equal(t, r.dirname, "/bar")
}

func Test_Reader_String(t *testing.T) {
	r, err := New("/foo/../bar/", "pattern", AsPlain)
	assert.NoError(t, err)
	// test that return "/bar/pattern"
	assert.Equal(t, r.String(), "/bar/pattern")
}

func Test_Reader_MatchString(t *testing.T) {
	// test that compare in plain text
	r, err := New("/foo/../bar/", `.*.tar.gz`, AsPlain)
	assert.NoError(t, err)
	assert.True(t, r.MatchString(".*.tar.gz"))
	assert.False(t, r.MatchString("test-match.tar.gz"))

	// test that matches the specified plain text
	r, err = New("/foo/../bar/", `.*.tar.gz`, AsRegex)
	assert.NoError(t, err)
	assert.True(t, r.MatchString(".*.tar.gz"))
	assert.True(t, r.MatchString("test-match.tar.gz"))

	// test that matches the specified plain text
	r, err = New("/foo/../bar/", `.*.tar.gz`, AsPosix)
	assert.NoError(t, err)
	assert.True(t, r.MatchString(".*.tar.gz"))
	assert.True(t, r.MatchString("test-match.tar.gz"))
}
