package telemetry_test

import (
	"testing"

	"github.com/nimona/go-nimona/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestCollectSuccess(t *testing.T) {
	ic, err := telemetry.NewInfluxCollector(
		"root", "", "http://localhost:8086")
	assert.NoError(t, err)

	for i := 0; i <= 2000; i++ {
		ic.Collect(&telemetry.ConnectionEvent{Outgoing: true})
	}

}
