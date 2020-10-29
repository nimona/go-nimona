package log

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TestNestedNamedOutput(t *testing.T) {
	r0, w0, _ := os.Pipe()
	r1, w1, _ := os.Pipe()

	DefaultLogLevel = DebugLevel

	logger := New()
	logger.SetOutput(w0)
	logger.Debug("foo")
	logger.Info("foo")
	logger.Warn("foo")

	logger = logger.Named("bar")
	logger.Error("foo")

	logger = logger.Named("bar2")
	logger.SetOutput(w1)
	logger.Error("foo")

	w0.Close()
	w1.Close()

	out0, _ := ioutil.ReadAll(r0) // nolint: errcheck
	out1, _ := ioutil.ReadAll(r1) // nolint: errcheck
	lines0 := strings.Split(strings.TrimSpace(string(out0)), "\n")
	lines1 := strings.Split(strings.TrimSpace(string(out1)), "\n")
	assert.Len(t, lines0, 4)
	assert.Len(t, lines1, 1)
}
