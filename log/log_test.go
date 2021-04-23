package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Error(t *testing.T) {
	w := Stderr
	b := bytes.NewBuffer(nil)
	Stderr = b
	defer func() {
		Stderr = w
	}()

	// test that output arguments to Stderr
	b.Reset()
	s := "format %q"
	Error(s, "hello", 1, []string{})
	assert.Equal(t, "format %q hello 1 []\n", b.String())

}

func Test_Print(t *testing.T) {
	w := Stdout
	b := bytes.NewBuffer(nil)
	Stdout = b
	defer func() {
		Stdout = w
	}()

	// test that output arguments to Stdout
	b.Reset()
	s := "format %q"
	Print(s, "hello", 1, []string{})
	assert.Equal(t, "format %q hello 1 []\n", b.String())
}

func Test_Fatal(t *testing.T) {
	w := Stderr
	b := bytes.NewBuffer(nil)
	Stderr = b

	exitfn := exit
	var exitCode int
	exit = func(code int) {
		exitCode = code
	}

	defer func() {
		Stderr = w
		exit = exitfn
	}()

	// test that output arguments to Stderr then call osExit(1)
	b.Reset()
	exitCode = 0
	s := "format %q"
	Fatal(s, "hello", 1, []string{})
	assert.Equal(t, "format %q hello 1 []\n", b.String())
	assert.Equal(t, 1, exitCode)
}

func Test_Errorf(t *testing.T) {
	w := Stderr
	b := bytes.NewBuffer(nil)
	Stderr = b

	// test that output formatted-string to Stderr
	Errorf("format %q", "hello")
	Stderr = w
	assert.Equal(t, "format \"hello\"\n", b.String())
}

func Test_Printf(t *testing.T) {
	w := Stdout
	b := bytes.NewBuffer(nil)
	Stdout = b

	// test that output formatted-string to Stdout
	Printf("format %q", "hello")
	Stdout = w
	assert.Equal(t, "format \"hello\"\n", b.String())
}

func Test_Fatalf(t *testing.T) {
	w := Stderr
	b := bytes.NewBuffer(nil)
	Stderr = b

	exitfn := exit
	var exitCode int
	exit = func(code int) {
		exitCode = code
	}

	// test that output formatted-string to Stderr then call osExit(1)
	Fatalf("format %q", "hello")
	Stderr = w
	exit = exitfn
	assert.Equal(t, "format \"hello\"\n", b.String())
	assert.Equal(t, 1, exitCode)
}
