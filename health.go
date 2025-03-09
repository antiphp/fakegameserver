package fakegameserver

import (
	"context"
	"errors"
	"maps"
	"slices"

	"github.com/antiphp/fakegameserver/agones"
)

const (
	// MessageTypeHealthStatus is the message type for health status updates.
	MessageTypeHealthStatus MessageType = "healthStatus"
)

var (
	_ Producer = (*HealthStatus)(nil)
	_ Consumer = (*HealthStatus)(nil)
)

// HealthStatus consumes all messages and reports the overall health status of the game server.
type HealthStatus struct {
	client   *agones.Client
	waitFor  []MessageType
	excludes []MessageType
	ch       chan map[MessageType]error
}

// NewHealthStatus returns a new health status reporter.
func NewHealthStatus() *HealthStatus {
	return &HealthStatus{
		ch: make(chan map[MessageType]error, 1),
	}
}

// Exclude excludes message types from the health status report.
func (r *HealthStatus) Exclude(types ...MessageType) {
	r.excludes = append(r.excludes, types...)
}

// WaitFor waits for message types to be received before reporting the health status.
func (r *HealthStatus) WaitFor(types ...MessageType) {
	r.waitFor = append(r.waitFor, types...)
}

// Close closes the health status reporter.
func (r *HealthStatus) Close() {
	close(r.ch)
}

// Run runs the health status reporter.
func (r *HealthStatus) Run(ctx context.Context, queue Queue) {
	health := make(map[MessageType]error)
	first := true
	was := false
	for {
		var report map[MessageType]error
		select {
		case <-ctx.Done():
			return
		case report = <-r.ch:
		}

		for typ, err := range report {
			r.waitFor = slices.DeleteFunc(r.waitFor, func(t MessageType) bool {
				return t == typ && err == nil
			})

			health[typ] = err
		}

		if len(r.waitFor) > 0 {
			continue
		}

		is := isHealthy(health)

		switch {
		case is && (first || !was):
			queue.Add(Message{
				Type:        MessageTypeHealthStatus,
				Description: "Game server became healthy",
				Payload:     true,
			})
		case !is && (first || was):
			queue.Add(Message{
				Type:        MessageTypeHealthStatus,
				Description: "Game server became unhealthy",
				Error:       errors.Join(slices.Collect(maps.Values(health))...),
				Payload:     false,
			})
		}

		first = false
		was = is
	}
}

// Consume consumes all messages.
func (r *HealthStatus) Consume(msg Message) Message {
	if slices.Contains(r.excludes, msg.Type) {
		return msg
	}

	r.ch <- map[MessageType]error{
		msg.Type: msg.Error,
	}
	return msg
}

// isHealthy returns true if all reports are healthy.
func isHealthy(reports map[MessageType]error) bool {
	for _, err := range reports {
		if err != nil {
			return false
		}
	}
	return true
}
