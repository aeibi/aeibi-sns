package auth

import (
	"context"
	"strings"

	"aeibi/util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const metadataAuthorizationKey = "authorization"

func NewAuthUnaryServerInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		accessToken := ""
		for _, authHeader := range md.Get(metadataAuthorizationKey) {
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				accessToken = strings.TrimSpace(authHeader[7:])
			}
		}
		claims, err := util.ParseJWT(accessToken, secret)
		authInfo := AuthInfo{
			Subject: "",
			Object:  info.FullMethod,
			Action:  "CALL",
		}
		if err == nil && claims != nil {
			authInfo.Subject = claims.Subject
		}
		// TODO: Casbin Auth
		ctx = WithAuthInfo(ctx, authInfo)
		return handler(ctx, req)
	}
}
