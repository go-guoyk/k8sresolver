package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"net"
	"os"
)

type server struct {
}

func (s *server) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (s *server) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdin})

	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, &server{})

	var l net.Listener
	var err error

	if l, err = net.Listen("tcp", ":5000"); err != nil {
		panic(err)
	}
	defer l.Close()

	s.Serve(l)
}
