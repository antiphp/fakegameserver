package fakegameserver

import (
	"context"
	"time"
)

type ExitTimer struct {
	dur time.Duration
}

func NewExitTimer(dur time.Duration) *ExitTimer {
	return &ExitTimer{
		dur: dur,
	}
}

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
