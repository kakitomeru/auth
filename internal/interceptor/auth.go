package interceptor

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthUnaryServerInterceptor(
	protectedSuffixes []string,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		isProtected := false
		for _, suffix := range protectedSuffixes {
			if strings.HasSuffix(info.FullMethod, suffix) {
				isProtected = true
				break
			}
		}

		if !isProtected {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
		}

		userIDValues := md.Get("x-user-id")
		if len(userIDValues) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
		}
		userIDStr := userIDValues[0]

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
		}

		ctx = context.WithValue(ctx, userIDKey, userID)

		return handler(ctx, req)
	}
}

func GetUserID(ctx context.Context) uuid.UUID {
	userID := ctx.Value(userIDKey).(uuid.UUID)
	return userID
}
