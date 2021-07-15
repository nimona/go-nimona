package keystream

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xujiajun/nutsdb"
)

func TestController_New(t *testing.T) {
	opt := nutsdb.DefaultOptions
	opt.Dir = "/tmp/nutsdb"
	db, err := nutsdb.Open(opt)
	require.NoError(t, err)
	// nolint: gocritic
	defer db.Close()
}
