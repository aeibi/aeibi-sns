package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// ServiceRegistrar wires a single API into both gRPC and the HTTP gateway.
type ServiceRegistrar struct {
	Name            string
	RegisterGRPC    func(*grpc.Server)
	RegisterGateway func(context.Context, *runtime.ServeMux) error
}
