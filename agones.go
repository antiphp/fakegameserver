package fakegameserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agones.dev/agones/pkg/sdk"
	"github.com/cenkalti/backoff/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UpdateState struct {
	State    State
	UpdateFn func(context.Context) error
}

type State string

const (
	StateReady     State = "Ready"
	StateAllocated State = "Allocated"
	StateShutdown  State = "Shutdown"
)

type Agones struct {
	client sdk.SDKClient

	states []State
	durs   []time.Duration

	mu       sync.Mutex
	listener []chan<- *sdk.GameServer

	wg sync.WaitGroup
	c  chan *sdk.GameServer
}

func NewAgonesClient(addr string) (sdk.SDKClient, error) {
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

func NewAgones(client sdk.SDKClient) *Agones {
	return &Agones{
		client: client,
		c:      make(chan *sdk.GameServer),
	}
}

func (a *Agones) Close() {
	a.wg.Wait()
	close(a.c)
}

func (a *Agones) Run(ctx context.Context) {
	a.wg.Add(1)
	defer a.wg.Done()

	go a.watchGameServer(ctx)

	for {
		var gs *sdk.GameServer
		select {
		case <-ctx.Done():
			return
		case gs = <-a.c:
		}

		func() {
			a.mu.Lock()
			defer a.mu.Unlock()

			for _, c := range a.listener {
				select {
				case <-ctx.Done():
				case c <- gs:
				}
			}
		}()
	}
}

func (a *Agones) UpdateState(ctx context.Context, st State) error {
	switch st {
	case StateReady:
		_, err := a.client.Ready(ctx, &sdk.Empty{})
		return err
	case StateAllocated:
		_, err := a.client.Allocate(ctx, &sdk.Empty{})
		return err
	case StateShutdown:
		_, err := a.client.Shutdown(ctx, &sdk.Empty{})
		return err
	default:
		return fmt.Errorf("unknown state: %s", st)
	}
}

func (a *Agones) WatchStateChanged(ctx context.Context) <-chan State {
	watchCh := make(chan *sdk.GameServer)
	go func() {
		defer close(watchCh)
		<-ctx.Done()
	}()

	a.mu.Lock()
	a.listener = append(a.listener, watchCh)
	a.mu.Unlock()

	ch := make(chan State)
	go func() {
		defer close(ch)

		var gs *sdk.GameServer
		var prevState State
		for {
			select {
			case <-ctx.Done():
				return
			case gs = <-watchCh:
			}

			state := State(gs.Status.State)
			if state == prevState {
				continue
			}

			if prevState != "" { // Skip first change, usually not interesting.
				select {
				case <-ctx.Done():
				case ch <- state:
				}
			}

			prevState = state
		}
	}()
	return ch
}

func (a *Agones) watchGameServer(ctx context.Context) {
	a.wg.Add(1)
	defer a.wg.Done()

	bo := backoff.NewExponentialBackOff()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(bo.NextBackOff()):
		}

		conn, err := a.client.WatchGameServer(ctx, &sdk.Empty{})
		if err != nil {
			continue
		}
		bo.Reset()

		for {
			gs, err := conn.Recv()
			if err != nil {
				break
			}

			select {
			case <-ctx.Done():
			case a.c <- gs:
			}
		}
	}
}
