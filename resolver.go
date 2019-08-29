package k8sresolver

import (
	"context"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

type Resolver struct {
	target Target
	conn   resolver.ClientConn
	opt    resolver.BuildOption
	client *Client

	cancel   context.CancelFunc
	resolves chan interface{}
}

func NewResolver(target Target, cc resolver.ClientConn, opts resolver.BuildOption, client *Client) *Resolver {
	r := &Resolver{
		target: target,
		conn:   cc,
		opt:    opts,
		client: client,

		resolves: make(chan interface{}, 1),
	}
	return r
}

func (r *Resolver) Start() {
	if r.cancel != nil {
		return
	}
	var ctx context.Context
	ctx, r.cancel = context.WithCancel(context.Background())
	go r.run(ctx)
}

func (r *Resolver) run(ctx context.Context) {
	vals := make(chan []string, 1)
	errs := make(chan error, 1)

	// watch
	go r.client.WatchAddress(ctx, r.target, vals, errs)

	// periodically update addresses
	tk := time.NewTicker(time.Second * 30)
	defer tk.Stop()

	// initial resolve
	r.resolves <- nil

	// last time resolved
	var lastResolved time.Time

	for {
		select {
		case addrs := <-vals:
			// on addresses updated
			r.updateAddresses(addrs)
		case err := <-errs:
			// on error occurred
			log.Printf("WatchAddresses Error: %s", err.Error())
		case <-tk.C:
			// on ticked
			r.resolves <- nil
			continue
		case <-r.resolves:
			// on resolve requested
			// de-duplicate requests
			if time.Since(lastResolved) < time.Second*5 {
				continue
			}
			lastResolved = time.Now()
			// resolve
			if addrs, err := r.client.GetAddresses(ctx, r.target); err != nil {
				log.Printf("GetAddresses Error: %s", err.Error())
				continue
			} else {
				r.updateAddresses(addrs)
			}
		case <-ctx.Done():
			// on closed
			return
		}
	}
}

func (r *Resolver) updateAddresses(addrs []string) {
	state := resolver.State{}
	for _, addr := range addrs {
		state.Addresses = append(state.Addresses, resolver.Address{Addr: addr, Type: resolver.Backend})
	}
	r.conn.UpdateState(state)
}

func (r *Resolver) resolveNow() {
	r.resolves <- nil
}

func (r *Resolver) ResolveNow(opt resolver.ResolveNowOption) {
	log.Printf("ResolveNow() called: %#v", opt)
	go r.resolveNow()
}

func (r *Resolver) Close() {
	if r.cancel == nil {
		return
	}
	r.cancel()
	r.cancel = nil
}
