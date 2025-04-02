package fakegameserver

import (
	"context"
	"time"
)

// ExitTimer is a timer that sends an exit message to the queue when it elapses.
type ExitTimer struct {
	dur time.Duration
}

// NewExitTimer creates a new exit timer with the specified duration.
func NewExitTimer(dur time.Duration) *ExitTimer {
	return &ExitTimer{
		dur: dur,
	}
}

// Run starts the exit timer and sends a message to the queue when the timer elapses.
func (e *ExitTimer) Run(ctx context.Context, q Queue) {
	q.Add(Message{
		Type:        MessageTypeInfo,
		Description: "Exit timer started with " + e.dur.String(),
	})

	select {
	case <-ctx.Done():
		return
	case <-time.After(e.dur):
		q.Add(Message{
			Type:        MessageTypeExit,
			Description: "Exit timer elapsed after " + e.dur.String(),
		})
	}
}
