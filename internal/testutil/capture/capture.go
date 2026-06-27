// Package capture provides helpers for capturing stdout in tests.
package capture

import (
	"bytes"
	"io"
	"os"
)

// Output runs fn and returns everything it wrote to stdout.
func Output(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic("capture.Output: failed to create pipe: " + err.Error())
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
