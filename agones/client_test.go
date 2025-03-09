package agones_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"agones.dev/agones/pkg/sdk"
	"agones.dev/agones/pkg/sdkserver"
	"github.com/antiphp/fakegameserver/agones"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClient_UpdateStateReady(t *testing.T) {
	tests := []struct {
		name        string
		sdkFuncName string
		state       agones.State
		err         error
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name:        "handles Ready",
			sdkFuncName: "Ready",
			state:       agones.StateReady,
			wantErr:     require.NoError,
		},
		{
			name:        "handles Ready error",
			sdkFuncName: "Ready",
			state:       agones.StateReady,
			err:         errors.New("test"),
			wantErr:     require.Error,
		},
		{
			name:        "handles Allocated",
			sdkFuncName: "Allocate",
			state:       agones.StateAllocated,
			wantErr:     require.NoError,
		},
		{
			name:        "handles Allocated error",
			sdkFuncName: "Allocate",
			state:       agones.StateAllocated,
			err:         errors.New("test"),
			wantErr:     require.Error,
		},
		{
			name:        "handles Shutdown",
			sdkFuncName: "Shutdown",
			state:       agones.StateShutdown,
			wantErr:     require.NoError,
		},
		{
			name:        "handles Shutdown error",
			sdkFuncName: "Shutdown",
			state:       agones.StateShutdown,
			err:         errors.New("test"),
			wantErr:     require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			m := &mockSDK{}
			m.On(test.sdkFuncName, &sdk.Empty{}).Return(&sdk.Empty{}, test.err).Once()

			client := agones.NewClient(m)
			err := client.UpdateState(t.Context(), test.state)

			test.wantErr(t, err)
			m.AssertExpectations(t)
		})
	}
}

func TestClient_WatchGameServer(t *testing.T) {
	sdkSrv, err := sdkserver.NewLocalSDKServer("", "fakeGameServer")
	require.NoError(t, err)
	t.Cleanup(sdkSrv.Close)

	sdkClient, err := agones.NewSDKClient("localhost:9357")
	require.NoError(t, err)

	client := agones.NewClient(sdkClient)
	go client.Run(t.Context())

	cond := sync.NewCond(&sync.Mutex{})
	var states []agones.State
	go client.WatchState(t.Context(), func(state agones.State) {
		cond.L.Lock()
		defer cond.L.Unlock()

		states = append(states, state)
		cond.Signal()
	})

	cond.L.Lock()
	_, _ = sdkClient.Ready(t.Context(), &sdk.Empty{})
	cond.Wait()
	_, _ = sdkClient.Allocate(t.Context(), &sdk.Empty{})
	cond.Wait()
	_, _ = sdkClient.Shutdown(t.Context(), &sdk.Empty{})
	cond.Wait()
	cond.L.Unlock()

	assert.Equal(t, []agones.State{
		agones.StateReady,
		agones.StateAllocated,
		agones.StateShutdown,
	}, states)
}

type mockSDK struct {
	mock.Mock
	mockSDKUnimplemented
}

func (m *mockSDK) Ready(_ context.Context, in *sdk.Empty, _ ...grpc.CallOption) (*sdk.Empty, error) {
	args := m.Called(in)
	return args.Get(0).(*sdk.Empty), args.Error(1)
}

func (m *mockSDK) Allocate(_ context.Context, in *sdk.Empty, _ ...grpc.CallOption) (*sdk.Empty, error) {
	args := m.Called(in)
	return args.Get(0).(*sdk.Empty), args.Error(1)
}

func (m *mockSDK) Shutdown(_ context.Context, in *sdk.Empty, _ ...grpc.CallOption) (*sdk.Empty, error) {
	args := m.Called(in)
	return args.Get(0).(*sdk.Empty), args.Error(1)
}

type mockSDKUnimplemented struct{}

func (m *mockSDKUnimplemented) Ready(context.Context, *sdk.Empty, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) Allocate(context.Context, *sdk.Empty, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) Shutdown(context.Context, *sdk.Empty, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) Health(context.Context, ...grpc.CallOption) (sdk.SDK_HealthClient, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) GetGameServer(context.Context, *sdk.Empty, ...grpc.CallOption) (*sdk.GameServer, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) WatchGameServer(context.Context, *sdk.Empty, ...grpc.CallOption) (sdk.SDK_WatchGameServerClient, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) SetLabel(context.Context, *sdk.KeyValue, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) SetAnnotation(context.Context, *sdk.KeyValue, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}

func (m *mockSDKUnimplemented) Reserve(context.Context, *sdk.Duration, ...grpc.CallOption) (*sdk.Empty, error) {
	panic("not implemented")
}
