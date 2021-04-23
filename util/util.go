package util

import (
	"runtime"
	"strings"
	"syscall"
)

var goexit = runtime.Goexit
var ExitCode int

func Exit(code int) {
	ExitCode = code
	goexit()
}

func Getenv(k string) (string, bool) {
	v, found := syscall.Getenv(k)
	if found {
		return strings.TrimSpace(v), true
	}
	return "", false
}
