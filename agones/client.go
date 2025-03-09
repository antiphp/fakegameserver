package agones

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"agones.dev/agones/pkg/sdk"
	"github.com/cenkalti/backoff/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// State represents an Agones state.
type State string

const (
	// StateReady is the Agones state Ready.
	// The state indicates that the game server is ready to receive traffic.
	StateReady State = "Ready"

	// StateAllocated is the Agones state Allocated.
	// The state indicates that the game server hosts a game session.
	StateAllocated State = "Allocated"

	// StateShutdown is the Agones state Shutdown.
	// The state indicates that the game server is shutting down.
	StateShutdown State = "Shutdown"
)

// Client is the Agones client.
type Client struct {
	client sdk.SDKClient
	health sdk.SDK_HealthClient

	isLocal atomic.Bool

	mu            sync.Mutex
	stateWatchers []func(State)
	connWatchers  []func(error)

	state   State
	connErr *error
}

// NewSDKClient returns a new Agones SDK client.
func NewSDKClient(addr string) (sdk.SDKClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(10*time.Second),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dialing Agones %s: %w", addr, err)
	}

	return sdk.NewSDKClient(conn), nil
}

// NewClient returns a new Agones client.
func NewClient(sdk sdk.SDKClient) *Client {
	return &Client{
		client: sdk,
	}
}

// IsLocal determines if the SDK server runs in local development mode.
func (c *Client) IsLocal() bool {
	return c.isLocal.Load()
}

// Health sends a health report.
func (c *Client) Health(ctx context.Context) error {
	if c.health == nil {
		var err error
		c.health, err = c.client.Health(ctx)
		if err != nil {
			return err
		}
	}
	if err := c.health.Send(&sdk.Empty{}); err != nil {
		c.health = nil // Try re-connect.
		return err
	}
	return nil
}

// UpdateState updates the state.
func (c *Client) UpdateState(ctx context.Context, st State) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch st {
	case StateReady:
		_, err := c.client.Ready(ctx, &sdk.Empty{})
		if err != nil {
			return fmt.Errorf("updating state to ready: %w", err)
		}
	case StateAllocated:
		_, err := c.client.Allocate(ctx, &sdk.Empty{})
		if err != nil {
			return fmt.Errorf("updating state to allocated: %w", err)
		}
	case StateShutdown:
		_, err := c.client.Shutdown(ctx, &sdk.Empty{})
		if err != nil {
			return fmt.Errorf("updating state to shutdown: %w", err)
		}
	default:
		return errors.New("unknown state: " + string(st))
	}
	return nil
}

// WatchConnection calls the given function when the Agones connectivity changes.
func (c *Client) WatchConnection(ctx context.Context, fn func(error)) {
	idx := c.subConnWatcher(fn)
	defer c.unsubConnWatcher(idx)

	<-ctx.Done()
}

// WatchState calls the given function when the Agones state changes.
func (c *Client) WatchState(ctx context.Context, fn func(State)) {
	idx := c.subStateWatcher(fn)
	defer c.unsubStateWatcher(idx)

	<-ctx.Done()
}

// Run runs the Agones client.
func (c *Client) Run(ctx context.Context) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	bo.MaxInterval = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(bo.NextBackOff()):
		}

		conn, err := c.client.WatchGameServer(ctx, &sdk.Empty{})
		if err != nil {
			c.notifyConnWatchers(err)
			continue
		}

		c.notifyConnWatchers(nil)
		bo.Reset()

		var raw *sdk.GameServer
		for {
			raw, err = conn.Recv()
			if err != nil {
				c.notifyConnWatchers(err)
				break
			}

			c.notifyStateWatchers(State(raw.GetStatus().GetState()))
			c.isLocal.Store(raw.GetObjectMeta().GetLabels()["islocal"] == "true")
		}
	}
}

func (c *Client) notifyConnWatchers(err error) {
	if c.connErr != nil && errors.Is(*c.connErr, err) {
		return
	}
	c.connErr = &err

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, fn := range c.connWatchers {
		fn(err)
	}
}

func (c *Client) notifyStateWatchers(state State) {
	if c.state == state {
		return
	}
	c.state = state

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, fn := range c.stateWatchers {
		fn(state)
	}
}

func (c *Client) subConnWatcher(fn func(error)) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connWatchers = append(c.connWatchers, fn)
	return len(c.connWatchers) - 1
}

func (c *Client) unsubConnWatcher(idx int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Overwrite instead of delete to keep indexes valid.
	c.connWatchers[idx] = func(error) {}
}

func (c *Client) subStateWatcher(fn func(State)) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stateWatchers = append(c.stateWatchers, fn)
	return len(c.stateWatchers) - 1
}

func (c *Client) unsubStateWatcher(idx int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Overwrite instead of delete to keep indexes valid.
	c.stateWatchers[idx] = func(State) {}
}
