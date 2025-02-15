package fakegameserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

type OptFunc func(*Server)

func WithExitAfter(dur time.Duration) OptFunc {
	return func(s *Server) {
		s.exitAfter = dur
	}
}

func WithUpdateStateAfter(state State, dur time.Duration, agones *Agones) OptFunc {
	return func(s *Server) {
		s.states = append(s.states, state)
		s.stateDurs = append(s.stateDurs, dur)
		s.agones = agones
	}
}

type Server struct {
	exitAfter time.Duration

	states    []State
	stateDurs []time.Duration
	agones    *Agones

	log *logger.Logger
}

func New(log *logger.Logger, opts ...OptFunc) *Server {
	s := Server{
		log: log,
	}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (s *Server) Run(ctx context.Context) (string, error) {
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

	var initStateChange chan int
	var changeState chan State
	var stateChanged <-chan State
	if len(s.states) > 0 {
		initStateChange = make(chan int, 1)
		defer close(initStateChange)

		initStateChange <- 0

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

		case i := <-initStateChange:
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
				case initStateChange <- i + 1:
				}
			}()
		}
	}
}
