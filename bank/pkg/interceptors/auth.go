package interceptors

import (
	"context"
	"fmt"
	"main/token"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	AuthorizationHeaderKey  string     = "authorization"
	AuthorizationType       string     = "bearer"
	AuthorizationPayloadKey contextKey = "authorization_payload"
)

func getgRPCRoutes() map[string][]string {
	const simpleBankServicesPath = "/pb.SimpleBank/"
	return map[string][]string{
		simpleBankServicesPath + "UpdateMe": {"user"},
	}
}
func getGatewayRoutes() map[string][]string {
	return map[string][]string{
		"PUT /v1/users": {"user"},
	}
}

type AuthInterceptor struct {
	tokenMaker      token.Maker
	accessibleRoles map[string][]string
}

func NewGRPCInterceptor(tokenMaker token.Maker) *AuthInterceptor {
	return &AuthInterceptor{
		tokenMaker:      tokenMaker,
		accessibleRoles: getgRPCRoutes(),
	}
}
func NewGatewayInterceptor(tokenMaker token.Maker) *AuthInterceptor {
	return &AuthInterceptor{
		tokenMaker:      tokenMaker,
		accessibleRoles: getGatewayRoutes(),
	}
}
func (authInterceptor *AuthInterceptor) AuthMiddleware(ctx context.Context, grpcMux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := authInterceptor.AuthorizeGateway(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if payload != nil {
			ctx = context.WithValue(ctx, AuthorizationPayloadKey, payload)
			r = r.WithContext(ctx)
		}

		grpcMux.ServeHTTP(w, r)
	})
}

// Update the Unary interceptor to apply interceptor middlewares
func (authInterceptor *AuthInterceptor) Unary() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(authInterceptor.LoggerInterceptor(), authInterceptor.AuthInterceptor())
}

func (authInterceptor *AuthInterceptor) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		payload, err := authInterceptor.AuthorizeGRPC(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		if payload != nil {
			ctx = context.WithValue(ctx, AuthorizationPayloadKey, payload)
		}

		return handler(ctx, req)
	}
}

// func (authInterceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
// 	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

//		}
//	}

func (authInterceptor *AuthInterceptor) AuthorizeGRPC(ctx context.Context, fullMethod string) (*token.Payload, error) {
	allowedRoles, exists := authInterceptor.accessibleRoles[fullMethod]
	if !exists {
		return nil, nil // Public endpoint
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	return authInterceptor.verifyAuth(md["authorization"], allowedRoles)
}

func (authInterceptor *AuthInterceptor) AuthorizeGateway(r *http.Request) (*token.Payload, error) {
	fullMethod := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	allowedRoles, exists := authInterceptor.accessibleRoles[fullMethod]
	if !exists {
		return nil, nil // Public endpoint
	}

	authHeader := []string{r.Header.Get("Authorization")}
	return authInterceptor.verifyAuth(authHeader, allowedRoles)
}

func (authInterceptor *AuthInterceptor) verifyAuth(authHeader []string, allowedRoles []string) (*token.Payload, error) {
	if len(authHeader) < 1 {
		return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	fields := strings.Fields(authHeader[0])
	if len(fields) < 2 {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token format")
	}

	if strings.ToLower(fields[0]) != AuthorizationType {
		return nil, status.Errorf(codes.Unauthenticated, "unsupported authorization type %v", AuthorizationType)
	}

	token := fields[1]
	payload, err := authInterceptor.tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token verification failed")
	}

	for _, role := range allowedRoles {
		if role == string(payload.Role) {
			return payload, nil
		}
	}

	return nil, status.Errorf(codes.PermissionDenied, "user role '%s' does not have access to this resource", payload.Role)
}
