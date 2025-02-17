// Package fakegs contains the fake game server runtime code.
package fakegs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

// OptFunc is an option function to apply server options.
type OptFunc func(*Server)

// WithExitAfter applies the option to exit after a specified duration.
func WithExitAfter(dur time.Duration) OptFunc {
	return func(s *Server) {
		s.exitAfter = dur
	}
}

// WithUpdateStateAfter applies the option to update the Agones state after a specified duration.
func WithUpdateStateAfter(state State, dur time.Duration, agones *Agones) OptFunc {
	return func(s *Server) {
		s.states = append(s.states, state)
		s.stateDurs = append(s.stateDurs, dur)
		s.agones = agones
	}
}

// Server is the fake game server.
type Server struct {
	exitAfter time.Duration

	states    []State
	stateDurs []time.Duration
	agones    *Agones

	log *logger.Logger
}

// New returns a new fake game server.
func New(log *logger.Logger, opts ...OptFunc) *Server {
	s := Server{
		log: log,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

// Run runs the runtime routine for the fake game server and returns a stop reason or error.
func (s *Server) Run(ctx context.Context) (string, error) { //nolint:cyclop // Readability.
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup features.

	var exitAfter <-chan time.Time
	if s.exitAfter > 0 {
		t := time.NewTimer(s.exitAfter)
		defer t.Stop()

		exitAfter = t.C
	}

	var (
		startStateChangeTimer chan int // Index of a.states[] & a.stateDurs[].
		changeState           chan State
		stateChanged          <-chan State
	)
	if len(s.states) > 0 {
		startStateChangeTimer = make(chan int, 1)
		defer close(startStateChangeTimer)

		startStateChangeTimer <- 0

		changeState = make(chan State)
		defer close(changeState)

		stateChanged = s.agones.WatchStateChanged(ctx)
	}

	// Feature loop.

	for {
		select {
		case <-ctx.Done():
			return ctx.Err().Error(), nil

		case <-exitAfter:
			return "exit timer elapsed after " + s.exitAfter.String(), nil

		case state := <-stateChanged:
			s.log.Debug("Agones state change detected", lctx.Str("state", string(state)))

		case state := <-changeState:
			s.log.Debug("Updating Agones state", lctx.Str("state", string(state)))
			if err := s.agones.UpdateState(ctx, state); err != nil {
				return "", fmt.Errorf("updating Agones state: %w", err)
			}

		case i := <-startStateChangeTimer:
			if i >= len(s.states) {
				break
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				select {
				case <-ctx.Done():
					return
				case <-time.After(s.stateDurs[i]):
				}

				select {
				case <-ctx.Done():
				case changeState <- s.states[i]:
				}

				select {
				case <-ctx.Done():
				case startStateChangeTimer <- i + 1:
				}
			}()
		}
	}
}
