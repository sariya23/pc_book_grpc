package service

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtManager      *JWTManager
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtWanager *JWTManager, roles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager:      jwtWanager,
		accessibleRoles: roles,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		log.Println("--> unary intercepter: ", info.FullMethod)

		err = i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream intercepter: ", info.FullMethod)
		err := i.authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := i.accessibleRoles[method]
	// If not roles for this method, then it method available
	// for all users.
	if !ok {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "auth token in not present")
	}
	token := values[0]
	claim, err := i.jwtManager.Verify(token)
	if err != nil {
		return status.Error(codes.Unauthenticated, "token is invalid")
	}

	for _, role := range accessibleRoles {
		if role == claim.Role {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "no permissions to access this RPC")
}
