package k8s

import (
	"fmt"
	v1 "github.com/ericchiang/k8s/apis/core/v1"
	"net"
	"sort"
	"strconv"
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

func (t Target) ResolveEndpoints(ep v1.Endpoints) (addrs []string, err error) {
	for _, sub := range ep.GetSubsets() {
		// resolve port
		var resolvedPort string
		if t.PortIsFirst || t.PortIsName {
			for _, p := range sub.GetPorts() {
				if (t.PortIsFirst || p.GetName() == t.Port) &&
					(p.GetProtocol() == "TCP" || p.GetProtocol() == "") {
					resolvedPort = strconv.Itoa(int(p.GetPort()))
					break
				}
			}
			if len(resolvedPort) == 0 {
				err = fmt.Errorf("failed to resolve port, name=%s, first=%t", t.Port, t.PortIsFirst)
				return
			}
		} else {
			resolvedPort = t.Port
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
