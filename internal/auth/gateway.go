package auth

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

func GatewayMetadataExtractor(_ context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		md.Append(metadataAuthorizationKey, authHeader)
	}
	return md
}
