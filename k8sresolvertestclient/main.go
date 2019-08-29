package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "go.guoyk.net/k8sresolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"os"
	"time"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdin})

	var err error
	var conn *grpc.ClientConn
	if conn, err = grpc.Dial(os.Getenv("TARGET"), grpc.WithInsecure()); err != nil {
		panic(err)
	}
	client := grpc_health_v1.NewHealthClient(conn)

	for {
		var resp *grpc_health_v1.HealthCheckResponse
		if resp, err = client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{}); err != nil {
			log.Error().Err(err).Msg("failed")
		} else {
			log.Info().Interface("resp", resp).Msg("succeeded")
		}
		time.Sleep(time.Millisecond * 500)
	}
}
