package k8s

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
