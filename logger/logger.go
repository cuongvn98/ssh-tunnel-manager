package logger

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

func Print(s string) (int, error) {
	return fmt.Fprint(stdout, s)
}

func Printf(format string, args ...any) (int, error) {
	return fmt.Fprintf(stdout, format, args...)
}
