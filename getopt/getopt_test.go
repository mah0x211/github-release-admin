package getopt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestOption struct {
	argv  map[string][]string
	count int
	nstop int
}

func (o *TestOption) SetArg(v string) bool {
	if o.nstop == 0 || o.count < o.nstop {
		o.argv["arg"] = append(o.argv["arg"], v)
		o.count++
	}
	return o.nstop == 0 || o.count < o.nstop
}

func (o *TestOption) SetFlag(v string) bool {
	if o.nstop == 0 || o.count < o.nstop {
		o.argv["flag"] = append(o.argv["flag"], v)
		o.count++
	}
	return o.nstop == 0 || o.count < o.nstop
}

func (o *TestOption) SetKeyValue(k, v, arg string) bool {
	if o.nstop == 0 || o.count < o.nstop {
		o.argv["kv"] = append(o.argv["kv"], arg)
		o.count++
	}
	return o.nstop == 0 || o.count < o.nstop
}

func TestParse(t *testing.T) {
	o := &TestOption{
		argv: map[string][]string{},
	}

	args := []string{
		"arg1", "--flag-x", "--kv-x=v-x", "--flag-y", "arg2",
	}
	// test that parse arguments
	assert.Empty(t, Parse(o, args))
	assert.Equal(t, []string{"arg1", "arg2"}, o.argv["arg"])
	assert.Equal(t, []string{"--flag-x", "--flag-y"}, o.argv["flag"])
	assert.Equal(t, []string{"--kv-x=v-x"}, o.argv["kv"])

	// test that consume at least one argument
	o.argv = map[string][]string{}
	o.nstop = 1
	o.count = 1
	assert.Equal(t, []string{"--flag-x", "--kv-x=v-x", "--flag-y", "arg2"}, Parse(o, args))
	assert.Empty(t, o.argv["arg"])
	assert.Empty(t, o.argv["flag"])
	assert.Empty(t, o.argv["kv"])

	// test that parse 3 arguments
	o.count = 0
	o.nstop = 3
	assert.Equal(t, []string{"--flag-y", "arg2"}, Parse(o, args))
	assert.Equal(t, []string{"arg1"}, o.argv["arg"])
	assert.Equal(t, []string{"--flag-x"}, o.argv["flag"])
	assert.Equal(t, []string{"--kv-x=v-x"}, o.argv["kv"])
}
