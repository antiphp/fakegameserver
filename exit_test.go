//go:build goexperiment.synctest

package fakegameserver_test

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/antiphp/fakegameserver"
	"github.com/antiphp/fakegameserver/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitTimer_Run(t *testing.T) {
	synctest.Run(func() {
		q := queue.NewFifo[fakegameserver.Message]()
		t.Cleanup(q.Shutdown)

		exit := fakegameserver.NewExitTimer(time.Minute)
		go exit.Run(t.Context(), q)

		msg, shutdown := q.Get()

		require.False(t, shutdown)
		assert.Equal(t, fakegameserver.MessageTypeInfo, msg.Type)

		synctest.Wait()

		msg, shutdown = q.Get()

		require.False(t, shutdown)
		assert.Equal(t, fakegameserver.MessageTypeExit, msg.Type)
	})
}
