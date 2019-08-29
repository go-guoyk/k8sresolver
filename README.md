# k8sresolver

kubernetes resolver support for gRPC

Inspired by https://github.com/sercand/kuberesolver, using a better kubernetes client

## Usage

```go
import _ "go.guoyk.net/k8sresolver"

func main() {
  conn, err := grpc.Dial("k8s:///someservice.somenamespace:5000")
  _, _ = conn, err
}
```

## Permission

By default, Pod uses the `default` ServiceAccount, thus read permission of `endpoints` must be assigned to `default` ServiceAccount

This is an example, modify if needed

```yaml

---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: endpoints-ro
  namespace: default
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: default-sa-to-endpoints-ro
  namespace: default
subjects:
  - kind: ServiceAccount
    name: default
roleRef:
  kind: Role
  name: endpoints-ro
  apiGroup: rbac.authorization.k8s.io
```

## License

Guo Y.K., MIT License
