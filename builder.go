package k8sresolver

import (
	"fmt"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"strings"
)

func init() {
	resolver.Register(&Builder{})
}

type Builder struct {
}

func (k *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (r resolver.Resolver, err error) {
	var client *Client
	if client, err = GetClient(); err != nil {
		return
	}
	var tgt Target
	if tgt, err = ResolveTarget(target.Authority, target.Endpoint, client.GetNamespace()); err != nil {
		return
	}
	r = NewResolver(tgt, cc, opts, client)
	return
}

func (k *Builder) Scheme() string {
	return "k8s"
}

func ResolveTarget(srcAuthority string, srcEndpoint string, currentNamespace string) (Target, error) {
	// k8s://default/service:port
	ep := srcEndpoint
	sNS := srcAuthority
	// k8s://service.default:port/
	if ep == "" {
		ep = srcAuthority
		sNS = currentNamespace
	}
	// k8s:///service:port
	// k8s://service:port/
	if sNS == "" {
		sNS = currentNamespace
	}

	out := Target{}
	if ep == "" {
		return Target{}, fmt.Errorf("target(%s/%s) is empty", srcAuthority, srcEndpoint)
	}
	var name string
	var port string
	if strings.LastIndex(ep, ":") < 0 {
		name = ep
		port = ""
		out.PortIsFirst = true
	} else {
		var err error
		name, port, err = net.SplitHostPort(ep)
		if err != nil {
			return Target{}, fmt.Errorf("target endpoint='%s' is invalid. grpc target is %s/%s, err=%v", ep, srcAuthority, srcEndpoint, err)
		}
	}

	nameSplit := strings.SplitN(name, ".", 2)
	sName := name
	if len(nameSplit) == 2 {
		sName = nameSplit[0]
		sNS = nameSplit[1]
	}
	out.Service = sName
	out.Namespace = sNS
	out.Port = port
	if !out.PortIsFirst {
		if _, err := strconv.Atoi(out.Port); err != nil {
			out.PortIsName = true
		} else {
			out.PortIsName = false
		}
	}
	return out, nil
}
