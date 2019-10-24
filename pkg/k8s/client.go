package k8s

import (
	"context"
	"fmt"
	"github.com/ericchiang/k8s"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var (
	defaultClient      *client
	defaultClientError error
	defaultClientOnce  = &sync.Once{}
)

type Client interface {
	GetNamespace() string
	GetAddresses(ctx context.Context, target Target) ([]string, error)
	WatchAddress(ctx context.Context, target Target, output chan []string)
}

type client struct {
	client *k8s.Client
}

func GetClient() (Client, error) {
	defaultClientOnce.Do(func() {
		var kc *k8s.Client
		if kc, defaultClientError = k8s.NewInClusterClient(); defaultClientError != nil {
			return
		}
		defaultClient = &client{client: kc}
	})
	return defaultClient, defaultClientError
}

func (c *client) GetNamespace() string {
	return c.client.Namespace
}

func (c *client) GetAddresses(ctx context.Context, target Target) ([]string, error) {
	var ep v1.Endpoints

	if err := c.client.Get(ctx, target.Namespace, target.Service, &ep); err != nil {
		return nil, err
	}

	return target.ResolveEndpoints(ep)
}

func (c *client) watchAddresses(ctx context.Context, target Target, output chan []string) (err error) {
	var watcher *k8s.Watcher
	if watcher, err = c.client.Watch(
		ctx,
		target.Namespace,
		&v1.Endpoints{},
		k8s.QueryParam("fieldSelector", fmt.Sprintf("metadata.name=%s", target.Service)),
	); err != nil {
		return
	}
	defer watcher.Close()
	for {
		ep := &v1.Endpoints{}
		var event string
		if event, err = watcher.Next(ep); err != nil {
			return
		}
		log.Debug().Str("event", event).Interface("endpoints", ep.Subsets).Msg("k8s client event received")
		var addrs []string
		if addrs, err = target.ResolveEndpoints(*ep); err != nil {
			return
		}
		// prevent blocking
		select {
		case output <- addrs:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (c *client) WatchAddress(ctx context.Context, target Target, output chan []string) {
	for {
		if err := c.watchAddresses(ctx, target, output); err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				log.Error().Err(err).Msg("kubernetes: failed to watch endpoints")
				time.Sleep(time.Second)
			}
		}
		// return if cancelled or timed out
		if ctx.Err() != nil {
			return
		}
	}
}
