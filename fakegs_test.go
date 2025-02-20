package fakegs_test

import (
	"os"
	"testing"
	"time"

	"github.com/antiphp/fakegs"
	"github.com/hamba/logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_RunWithExitAfter(t *testing.T) {
	log := logger.New(os.Stdout, logger.LogfmtFormat(), logger.Info)

	tests := []struct {
		name      string
		exitAfter time.Duration
		tolerance time.Duration
	}{
		{
			name:      "exits immediately",
			exitAfter: 1, // 1ns
			tolerance: 100 * time.Millisecond,
		},
		{
			name:      "exits after 3s",
			exitAfter: 3 * time.Second,
			tolerance: 1 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := fakegs.New(log, fakegs.WithExitAfter(test.exitAfter))
			start := time.Now()

			reason, err := srv.Run(t.Context())

			require.NoError(t, err)
			assert.Contains(t, reason, "exit timer")
			assert.Contains(t, reason, test.exitAfter.String())
			assert.InDeltaf(t, test.exitAfter, time.Since(start), float64(2*time.Second), "Sanity check.")
		})
	}
}
