package k8sresolver

import (
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	target     Target
	clientConn resolver.ClientConn
	opts       resolver.BuildOption
	client     *Client
}

func NewResolver(target Target, cc resolver.ClientConn, opts resolver.BuildOption, client *Client) *Resolver {
	r := &Resolver{
		target:     target,
		clientConn: cc,
		opts:       opts,
		client:     client,
	}
	client.AddWatcher(target, r)
	return r
}

func (r *Resolver) HandleAddresses(target Target, addrs []string) {
	conn := r.clientConn
	if conn == nil {
		return
	}
	state := resolver.State{}
	for _, addr := range addrs {
		state.Addresses = append(state.Addresses, resolver.Address{Addr: addr, Type: resolver.Backend})
	}
	conn.UpdateState(state)
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOption) {
	r.client.RequestUpdate(r.target)
}

func (r *Resolver) Close() {
	r.client.RemoveWatcher(r)
}
