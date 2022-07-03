package log_terminal

import (
	"fmt"
	"os"

	"github.com/apoorvam/goterminal"
)

var stdout *goterminal.Writer

func init() {
	stdout = goterminal.New(os.Stdout)
}

func Reset() {
	stdout.Reset()
}

func Clear() {
	stdout.Clear()
}

func Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(stdout, format, args...)
}

func Show() {
	stdout.Print()
}
