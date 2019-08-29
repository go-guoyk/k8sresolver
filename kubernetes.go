package k8sresolver

import (
	"context"
	"fmt"
	"github.com/ericchiang/k8s"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	"net"
	"strconv"
	"sync"
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

type Client struct {
	client *k8s.Client
}

type Watcher interface {
	HandleAddresses(target Target, addrs []string)
}

func NewClient() (client *Client, err error) {
	var klient *k8s.Client
	if klient, err = k8s.NewInClusterClient(); err != nil {
		return
	}
	client = &Client{client: klient}
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

func (c *Client) GetAddresses(ctx context.Context, target Target) (addrs []string, err error) {
	var ep v1.Endpoints

	if err = c.client.Get(ctx, target.Namespace, target.Service, &ep); err != nil {
		return
	}

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
	return
}

func (c *Client) RequestUpdate(target Target) {
}

func (c *Client) AddWatcher(target Target, w Watcher) {
}

func (c *Client) RemoveWatcher(w Watcher) {
}
