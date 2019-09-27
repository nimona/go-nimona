package containers_test

import (
	"nimona.io/pkg/context"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/pkg/simulation/containers"
)

func TestNetwork(t *testing.T) {

	tests := []struct {
		name    string
		netType string
	}{
		{
			name:    "HappyPath",
			netType: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			n1, err := containers.NewNetwork(ctx, tt.name)
			assert.NoError(t, err)
			assert.NotNil(t, n1)
			defer func() {
				err := n1.Remove(ctx)
				assert.NoError(t, err)
			}()
		})
	}

}
