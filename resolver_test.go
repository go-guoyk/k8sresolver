package k8sresolver

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.guoyk.net/k8sresolver/pkg/k8s"
	"google.golang.org/grpc/resolver"
	"testing"
	"time"
)

type testClientConn struct {
	state         resolver.State
	addresses     []resolver.Address
	serviceConfig string
	updateCalled  int
}

func (t *testClientConn) UpdateState(state resolver.State) {
	t.updateCalled++
	t.state = state
}

func (t *testClientConn) NewAddress(addresses []resolver.Address) {
	t.addresses = addresses
}

func (t *testClientConn) NewServiceConfig(serviceConfig string) {
	t.serviceConfig = serviceConfig
}

type testK8sClient struct {
	target    k8s.Target
	getCalled int
}

func (t *testK8sClient) GetNamespace() string {
	return "default"
}

func (t *testK8sClient) GetAddresses(ctx context.Context, target k8s.Target) ([]string, error) {
	t.getCalled++
	t.target = target
	return []string{"1.1.1.1:80"}, nil
}

func (t *testK8sClient) WatchAddress(ctx context.Context, target k8s.Target, output chan []string) {
	t.target = target
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			output <- []string{"1.1.1.1:80"}
		case <-ctx.Done():
			return
		}
	}
}

func TestNewResolver(t *testing.T) {
	tg := k8s.Target{Namespace: "ns1", Service: "svc1", Port: "http"}
	cc := &testClientConn{}
	kc := &testK8sClient{}
	r := NewResolver(tg, cc, resolver.BuildOption{}, kc)
	r.Start()
	time.Sleep(time.Second)
	r.ResolveNow(resolver.ResolveNowOption{})
	time.Sleep(time.Second * 4)
	require.Equal(t, 1, cc.updateCalled)
	require.Equal(t, 2, kc.getCalled)
}
