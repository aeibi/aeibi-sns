package auth

import "context"

type contextKey string

const authInfoKey contextKey = "auth-info"

type AuthInfo struct {
	Subject string
	Object  string
	Action  string
}

func WithAuthInfo(ctx context.Context, info AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoKey, info)
}

func FromContext(ctx context.Context) (AuthInfo, bool) {
	if ctx == nil {
		return AuthInfo{}, false
	}
	val := ctx.Value(authInfoKey)
	if val == nil {
		return AuthInfo{}, false
	}

	info, ok := val.(AuthInfo)
	return info, ok
}

func SubjectFromContext(ctx context.Context) (string, bool) {
	info, ok := FromContext(ctx)
	if !ok || info.Subject == "" {
		return "", false
	}
	return info.Subject, true
}
