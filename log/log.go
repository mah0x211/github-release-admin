package log

import (
	"fmt"
	"io"
	"os"

	"github-release-admin/util"
)

var Stdout io.Writer = os.Stdout
var Stderr io.Writer = os.Stderr
var Verbose bool

func fprintln(w io.Writer, a ...interface{}) {
	fmt.Fprintln(w, a...)
}

func Error(a ...interface{}) {
	fprintln(Stderr, a...)
}

func Print(a ...interface{}) {
	fprintln(Stdout, a...)
}

var exit = util.Exit

func Fatal(a ...interface{}) {
	Error(a...)
	exit(1)
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	fmt.Fprintf(w, format+"\n", a...)
}

func Errorf(format string, a ...interface{}) {
	fprintf(Stderr, format, a...)
}

func Printf(format string, a ...interface{}) {
	fprintf(Stdout, format, a...)
}

func Debug(format string, a ...interface{}) {
	if Verbose {
		Printf(format, a...)
	}
}

func Fatalf(format string, a ...interface{}) {
	Errorf(format, a...)
	exit(1)
}
