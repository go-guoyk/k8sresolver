package k8sresolver

import (
	"context"
	"fmt"
	"github.com/ericchiang/k8s"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/rs/zerolog/log"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	defaultClient      *Client
	defaultClientError error
	defaultClientOnce  = &sync.Once{}
)

type Target struct {
	Service     string
	Namespace   string
	Port        string
	PortIsName  bool
	PortIsFirst bool
}

func (t Target) String() string {
	if t.PortIsFirst {
		return t.Service + "." + t.Namespace
	} else {
		return t.Service + "." + t.Namespace + ":" + t.Port
	}
}

type AddressesUpdate struct {
	Addrs []string
	Err   error
}

func endpointsToAddresses(ep v1.Endpoints, target Target) (addrs []string, err error) {
	for _, sub := range ep.GetSubsets() {
		// resolve port
		var resolvedPort string
		if target.PortIsFirst || target.PortIsName {
			for _, p := range sub.GetPorts() {
				if (target.PortIsFirst || p.GetName() == target.Port) &&
					(p.GetProtocol() == "TCP" || p.GetProtocol() == "") {
					resolvedPort = strconv.Itoa(int(p.GetPort()))
					break
				}
			}
			if len(resolvedPort) == 0 {
				err = fmt.Errorf("failed to resolve port, name=%s, first=%t", target.Port, target.PortIsFirst)
				return
			}
		} else {
			resolvedPort = target.Port
		}
		// add addresses
		for _, addr := range sub.Addresses {
			addrs = append(addrs, net.JoinHostPort(addr.GetIp(), resolvedPort))
		}
	}
	// sort strings for better LB
	sort.Strings(addrs)
	return
}

type Client struct {
	client *k8s.Client
}

func NewClient() (client *Client, err error) {
	var kc *k8s.Client
	if kc, err = k8s.NewInClusterClient(); err != nil {
		return
	}
	client = &Client{client: kc}
	return
}

func GetClient() (*Client, error) {
	defaultClientOnce.Do(func() {
		defaultClient, defaultClientError = NewClient()
	})
	return defaultClient, defaultClientError
}

func (c *Client) GetNamespace() string {
	return c.client.Namespace
}

func (c *Client) GetAddresses(ctx context.Context, target Target) ([]string, error) {
	var ep v1.Endpoints

	if err := c.client.Get(ctx, target.Namespace, target.Service, &ep); err != nil {
		return nil, err
	}

	return endpointsToAddresses(ep, target)
}

func (c *Client) watchAddresses(ctx context.Context, target Target, output chan []string) (err error) {
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
		if addrs, err = endpointsToAddresses(*ep, target); err != nil {
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

func (c *Client) WatchAddress(ctx context.Context, target Target, output chan []string) {
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
