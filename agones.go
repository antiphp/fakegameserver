package fakegameserver

import (
	"context"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/antiphp/fakegameserver/agones"
	"github.com/antiphp/fakegameserver/internal/exiterror"
	"k8s.io/utils/ptr"
)

const (
	// MessageTypeAgonesUpdate is the message type for Agones game server updates.
	MessageTypeAgonesUpdate MessageType = "agonesUpdate"

	// MessageTypeAgonesConnection is the message type for Agones connectivity updates.
	MessageTypeAgonesConnection MessageType = "agonesConnection"

	// MessageTypeAgonesReportHealth is the message type for health status reports.
	MessageTypeAgonesReportHealth MessageType = "agonesReportHealth"

	// MessageTypeAgonesRequestUpdate is the message type for Agones state update requests.
	MessageTypeAgonesRequestUpdate MessageType = "agonesRequestUpdate"
)

var _ Producer = (*AgonesWatcher)(nil)

// AgonesWatcher produces messages for any Agones connection or state update.
type AgonesWatcher struct {
	client *agones.Client
}

// NewAgonesWatcher returns a new Agones watcher.
func NewAgonesWatcher(client *agones.Client) *AgonesWatcher {
	return &AgonesWatcher{
		client: client,
	}
}

// Run runs the Agones watcher.
func (w *AgonesWatcher) Run(ctx context.Context, queue Queue) {
	first := true
	go w.client.WatchConnection(ctx, func(err error) {
		switch {
		case first && err != nil:
			queue.Add(Message{
				Type:        MessageTypeAgonesConnection,
				Description: "Agones connecting",
				Error:       err,
				Payload:     false, // Connected.
			})
		case err != nil:
			queue.Add(Message{
				Type:        MessageTypeAgonesConnection,
				Description: "Agones connection lost",
				Error:       err,
				Payload:     false,
			})
		default:
			queue.Add(Message{
				Type:        MessageTypeAgonesConnection,
				Description: "Agones connection established",
				Payload:     true,
			})
			first = false
		}
	})
	go w.client.WatchState(ctx, func(state agones.State) {
		queue.Add(Message{
			Type:        MessageTypeAgonesUpdate,
			Description: "Agones state change received for " + string(state),
			Payload:     state,
		})
	})
	<-ctx.Done()
}

var (
	_ Producer = (*AgonesStateUpdater)(nil)
	_ Consumer = (*AgonesStateUpdater)(nil)
)

// AgonesStateUpdater updates the Agones state when requested.
type AgonesStateUpdater struct {
	client  *agones.Client
	stateCh chan agones.State
}

// NewAgonesStateUpdater returns a new Agones state updater.
func NewAgonesStateUpdater(client *agones.Client) *AgonesStateUpdater {
	return &AgonesStateUpdater{
		client:  client,
		stateCh: make(chan agones.State, 1),
	}
}

// Run runs the Agones state updater.
func (u *AgonesStateUpdater) Run(ctx context.Context, queue Queue) {
	for {
		var state agones.State
		select {
		case <-ctx.Done():
			return
		case state = <-u.stateCh:
		}

		err := u.client.UpdateState(ctx, state)
		if err != nil {
			queue.Add(Message{
				Type:        MessageTypeAgonesUpdate,
				Description: "Agones state update failed",
				Error:       err,
				Payload:     state,
			})
			continue
		}

		queue.Add(Message{
			Type:        MessageTypeAgonesUpdate,
			Description: "Agones state updated",
			Payload:     state,
		})
	}
}

// Consume consumes Agones state update requests.
func (u *AgonesStateUpdater) Consume(msg Message) {
	if msg.Type != MessageTypeAgonesRequestUpdate {
		return
	}

	state, _ := msg.Payload.(agones.State)
	u.stateCh <- state
}

var (
	_ Producer = (*AgonesHealthReporter)(nil)
	_ Consumer = (*AgonesHealthReporter)(nil)
)

// AgonesHealthReporter reports the health of the game server to Agones by consuming health status messages.
type AgonesHealthReporter struct {
	client    *agones.Client
	initDelay time.Duration
	intvl     time.Duration

	ch chan bool
}

// NewAgonesHealthReporter returns a new Agones health reporter.
func NewAgonesHealthReporter(client *agones.Client, initDelay, intvl time.Duration) *AgonesHealthReporter {
	return &AgonesHealthReporter{
		client:    client,
		initDelay: initDelay,
		intvl:     intvl,
		ch:        make(chan bool, 1),
	}
}

// Run runs the Agones health reporter.
func (r *AgonesHealthReporter) Run(ctx context.Context, queue Queue) {
	start := time.Now()

	// For simplicity we are polling every second to determine what to do, but a reactive system would be nicer.
	t := time.NewTicker(time.Second)
	defer t.Stop()

	var (
		healthy  bool
		lastSent time.Time
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case healthy = <-r.ch:
		}

		if time.Since(lastSent) < r.intvl || !healthy || time.Since(start) < r.initDelay {
			continue
		}

		err := r.client.Health(ctx)
		switch {
		case err != nil:
			queue.Add(Message{
				Type:        MessageTypeAgonesReportHealth,
				Description: "Health report failed",
				Error:       err,
				Payload:     false,
			})
		default:
			queue.Add(Message{
				Type:        MessageTypeAgonesReportHealth,
				Description: "Health reported",
				Payload:     true,
			})

			lastSent = time.Now()
		}
	}
}

// Consume consumes health status messages.
func (r *AgonesHealthReporter) Consume(msg Message) {
	if msg.Type != MessageTypeHealthStatus {
		return
	}

	healthy, _ := msg.Payload.(bool)
	r.ch <- healthy
}

var (
	_ Producer = (*AgonesStateTimer)(nil)
	_ Consumer = (*AgonesStateTimer)(nil)
)

// AgonesStateTimer requests Agones state updates after configurable durations.
type AgonesStateTimer struct {
	states []agones.State
	durs   []time.Duration

	mu    sync.Mutex
	state agones.State

	once   sync.Once
	waitCh chan struct{}
}

// NewAgonesStateTimer returns a new Agones state timer.
func NewAgonesStateTimer() *AgonesStateTimer {
	return &AgonesStateTimer{
		waitCh: make(chan struct{}),
	}
}

// AddState adds a state and duration to the timer.
func (u *AgonesStateTimer) AddState(state agones.State, dur time.Duration) {
	u.states = append(u.states, state)
	u.durs = append(u.durs, dur)
}

// Run runs the Agones state timer.
func (u *AgonesStateTimer) Run(ctx context.Context, queue Queue) {
	select {
	case <-ctx.Done():
		return
	case <-u.waitCh:
	}

	states := slices.Clone(u.states)
	durs := slices.Clone(u.durs)
	var (
		state agones.State
		dur   time.Duration
	)
	for {
		if len(states) == 0 {
			return
		}

		state, states = shift(states)
		dur, durs = shift(durs)

		if u.getState() == state {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(dur):
		}

		queue.Add(Message{
			Type:        MessageTypeAgonesRequestUpdate,
			Description: "Requesting Agones state update to " + string(state),
			Payload:     state,
		})
	}
}

// Consume consumes Agones connection and state update messages.
func (u *AgonesStateTimer) Consume(msg Message) {
	switch {
	case msg.Type == MessageTypeAgonesConnection:
		if val, _ := msg.Payload.(bool); val {
			u.once.Do(func() { // Handle re-connects.
				close(u.waitCh)
			})
		}
	case msg.Type == MessageTypeAgonesUpdate:
		if state, ok := msg.Payload.(agones.State); ok {
			u.setState(state)
		}
	}
}

func (u *AgonesStateTimer) setState(state agones.State) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.state = state
}

func (u *AgonesStateTimer) getState() agones.State {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.state
}

func shift[T any](s []T) (T, []T) {
	if len(s) == 0 {
		var zero T
		return zero, nil
	}
	return s[0], s[1:]
}

// Shutdown is a shutdown handler that exits the game server when Agones state changes.
type Shutdown struct {
	enabledFn func() bool
	once      sync.Once
	waitCh    chan struct{}
}

// NewAgonesShutdown returns a new Agones shutdown handler.
func NewAgonesShutdown(enabledFn func() bool) *Shutdown {
	return &Shutdown{
		enabledFn: enabledFn,
		waitCh:    make(chan struct{}),
	}
}

// Run runs the shutdown handler.
func (s *Shutdown) Run(ctx context.Context, q Queue) {
	select {
	case <-ctx.Done():
		return
	case <-s.waitCh:
	}

	if !s.enabledFn() {
		return
	}

	q.Add(Message{
		Type:        MessageTypeExit,
		Description: "Agones state changed to Shutdown, emulating the behavior of Agones in a non-local development environment with SIGTERM",
		Error:       exiterror.New(nil, ptr.To[int](int(syscall.SIGTERM))),
	})
}

// Consume consumes Agones state update messages.
func (s *Shutdown) Consume(msg Message) {
	if msg.Type != MessageTypeAgonesUpdate {
		return
	}
	if state, ok := msg.Payload.(agones.State); ok && state != agones.StateShutdown {
		return
	}

	s.once.Do(func() {
		close(s.waitCh)
	})
}
