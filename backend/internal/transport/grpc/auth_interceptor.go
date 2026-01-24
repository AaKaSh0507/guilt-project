package grpc

import (
	"context"
	"strings"

	"guiltmachine/internal/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// publicMethods lists gRPC methods that don't require authentication
var publicMethods = map[string]bool{
	"/guiltmachine.v1.UserService/CreateUser":       true,
	"/guiltmachine.v1.UserService/GetUser":          true,
	"/guiltmachine.v1.SessionService/CreateSession": true,
}

// AuthInterceptor creates a unary interceptor for JWT authentication
func AuthInterceptor(jwtManager *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public methods
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract and validate token
		userID, sessionID, err := validateToken(ctx, jwtManager)
		if err != nil {
			return nil, err
		}

		// Add user info to context
		ctx = auth.ContextWithUserID(ctx, userID)
		ctx = auth.ContextWithSessionID(ctx, sessionID)

		return handler(ctx, req)
	}
}

// AuthStreamInterceptor creates a stream interceptor for JWT authentication
func AuthStreamInterceptor(jwtManager *auth.JWTManager) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip auth for public methods
		if publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		// Extract and validate token
		userID, sessionID, err := validateToken(stream.Context(), jwtManager)
		if err != nil {
			return err
		}

		// Wrap stream with authenticated context
		wrapped := &wrappedStream{
			ServerStream: stream,
			ctx:          auth.ContextWithSessionID(auth.ContextWithUserID(stream.Context(), userID), sessionID),
		}

		return handler(srv, wrapped)
	}
}

// validateToken extracts and validates JWT from context metadata
func validateToken(ctx context.Context, jwtManager *auth.JWTManager) (string, string, error) {
	token, err := extractToken(ctx)
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	userID, sessionID, err := jwtManager.Verify(token)
	if err != nil {
		return "", "", status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	return userID, sessionID, nil
}

// extractToken gets the JWT from the Authorization header
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// wrappedStream wraps a grpc.ServerStream with a custom context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
