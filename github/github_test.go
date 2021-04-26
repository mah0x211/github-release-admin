package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
