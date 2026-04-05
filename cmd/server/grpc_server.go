package server

import (
	"errors"
	"fmt"
	"net"

	"aeibi/internal/auth"
	"aeibi/internal/config"

	"google.golang.org/grpc"
)

// StartGRPCServer starts the gRPC server and returns it plus an error channel.
func StartGRPCServer(cfg *config.Config, registrars []ServiceRegistrar) (*grpc.Server, <-chan error, error) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.NewAuthUnaryServerInterceptor(cfg.Auth.JWTSecret)),
	)
	for _, registrar := range registrars {
		if registrar.RegisterGRPC != nil {
			registrar.RegisterGRPC(grpcServer)
		}
	}

	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("listen gRPC: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- err
		}
	}()

	return grpcServer, errCh, nil
}
