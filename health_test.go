package fakegameserver_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/antiphp/fakegameserver"
	"github.com/antiphp/fakegameserver/internal/queue"
	"github.com/hamba/testutils/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		excludes []fakegameserver.MessageType
		waitFor  []fakegameserver.MessageType
		consumes []fakegameserver.Message
		want     []bool
	}{
		{
			name: "handles becoming unhealthy",
			consumes: []fakegameserver.Message{
				{Type: "foo1"},
				{Type: "foo1", Error: errors.New("test")},
			},
			want: []bool{true, false},
		},
		{
			name: "handles becoming healthy",
			consumes: []fakegameserver.Message{
				{Type: "foo1", Error: errors.New("test")},
				{Type: "foo1"},
			},
			want: []bool{false, true},
		},
		{
			name: "handles multiple reporters",
			consumes: []fakegameserver.Message{
				{Type: "foo1", Error: errors.New("test")},
				{Type: "foo2"},
				{Type: "foo2", Error: errors.New("test")},
				{Type: "foo2"},
				{Type: "foo1"},
			},
			want: []bool{false, true},
		},
		{
			name:    "handles wait for",
			waitFor: []fakegameserver.MessageType{"foo2"},
			consumes: []fakegameserver.Message{
				{Type: "foo1"},
				{Type: "foo1", Error: errors.New("test")},
				{Type: "foo2"},
				{Type: "foo1"},
			},
			want: []bool{false, true},
		},
		{
			name:     "handles excludes",
			excludes: []fakegameserver.MessageType{"foo2"},
			consumes: []fakegameserver.Message{
				{Type: "foo1", Error: errors.New("test")},
				{Type: "foo1"},
				{Type: "foo1", Error: errors.New("test")},
				{Type: "foo2"},
				{Type: "foo1"},
			},
			want: []bool{false, true, false, true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := queue.NewFifo[fakegameserver.Message]()
			defer q.Shutdown()

			context.AfterFunc(t.Context(), func() {
				q.Shutdown()
			})

			health := fakegameserver.NewHealthStatus()
			health.Exclude(test.excludes...)
			health.WaitFor(test.waitFor...)
			go health.Run(t.Context(), q)

			for _, msg := range test.consumes {
				health.Consume(msg)
			}

			cond := sync.NewCond(&sync.Mutex{})
			var got []bool
			go func() {
				for {
					msg, shutdown := q.Get()
					if shutdown {
						break
					}

					require.Equal(t, fakegameserver.MessageTypeHealthStatus, msg.Type)

					cond.L.Lock()
					got = append(got, msg.Payload.(bool))
					cond.Signal()
					cond.L.Unlock()
				}
			}()

			retry.Run(t, func(t *retry.SubT) {
				cond.L.Lock()
				defer cond.L.Unlock()
				cond.Wait()

				assert.Equal(t, test.want, got)
			})

		})
	}
}
