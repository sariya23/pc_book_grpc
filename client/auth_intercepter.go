package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	authClient  *AuthClient
	authMethods map[string]bool
	token       string
}

func NewAuthIntercepter(
	ctx context.Context,
	authClient *AuthClient,
	authMethods map[string]bool,
	refreshDuration time.Duration,
) (*AuthInterceptor, error) {
	intercepter := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
	}
	err := intercepter.scheduleRefreshtoken(ctx, refreshDuration)
	if err != nil {
		return nil, err
	}
	return intercepter, nil
}

func (ai *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log.Printf("--> unary intercepter: %v\n", method)
		if ai.authMethods[method] {
			return invoker(ai.attachToken(ctx), method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (ai *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %v", method)
		if ai.authMethods[method] {
			return streamer(ai.attachToken(ctx), desc, cc, method, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (ai *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", ai.token)
}

func (ai *AuthInterceptor) scheduleRefreshtoken(ctx context.Context, refreshDuration time.Duration) error {
	err := ai.refreshToken(ctx)
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			time.Sleep(wait)
			err := ai.refreshToken(ctx)
			if err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()
	return nil
}

func (ai *AuthInterceptor) refreshToken(ctx context.Context) error {
	token, err := ai.authClient.Login(ctx)
	if err != nil {
		return err
	}
	ai.token = token
	log.Printf("token refreshed: %v", token)
	return nil
}
