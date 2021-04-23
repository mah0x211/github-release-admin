package util

import (
	"runtime"
)

var goexit = runtime.Goexit
var ExitCode int

func Exit(code int) {
	ExitCode = code
	goexit()
}
