package grpcserver

import (
	"context"
	"strings"

	"shortify/server/internal/auth"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(userIDKey).(uuid.UUID)
	return value, ok
}

func AuthInterceptor(tokens *auth.TokenManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 || !strings.HasPrefix(values[0], "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}

		claims, err := tokens.Parse(strings.TrimPrefix(values[0], "Bearer "))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		return handler(ctx, req)
	}
}

func isPublicMethod(method string) bool {
	switch method {
	case "/shortify.v1.AuthService/Register",
		"/shortify.v1.AuthService/Login",
		"/shortify.v1.LinkService/ResolveLink":
		return true
	default:
		return false
	}
}
