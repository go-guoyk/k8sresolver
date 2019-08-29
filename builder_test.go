package k8sresolver

import (
	"google.golang.org/grpc"
	"testing"
)

func TestBuilder_Build(t *testing.T) {
	conn, err := grpc.Dial("k8s://hello.default", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	_ = conn
}
