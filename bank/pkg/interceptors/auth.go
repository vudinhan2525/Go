package interceptors

import (
	"context"
	"fmt"
	"main/pkg/middlewares"
	"main/token"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func getgRPCRoutes() map[string][]string {
	const simpleBankServicesPath = "/pb.SimpleBank/"
	return map[string][]string{
		simpleBankServicesPath + "UpdateMe":   {"user", "admin"},
		simpleBankServicesPath + "CreateUser": {"admin"},
	}
}
func getGatewayRoutes() map[string][]string {
	return map[string][]string{
		"POST /v1/users": {"admin"},
		"PUT /v1/users":  {"admin", "user"},
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
func (authInterceptor *AuthInterceptor) GatewayMiddlewares(ctx context.Context, grpcMux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := authInterceptor.AuthorizeGateway(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		grpcMux.ServeHTTP(w, r)
	})
}

func (authInterceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		err = authInterceptor.AuthorizeGRPC(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// func (authInterceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
// 	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

//		}
//	}

func (authInterceptor *AuthInterceptor) AuthorizeGRPC(ctx context.Context, fullMethod string) error {
	allowedRoles, exists := authInterceptor.accessibleRoles[fullMethod]
	if !exists {
		// If method is not in the accessibleRoles map, assume it's public
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	return authInterceptor.verifyAuth(md["authorization"], allowedRoles)
}

func (authInterceptor *AuthInterceptor) AuthorizeGateway(r *http.Request) error {
	fullMethod := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	allowedRoles, exists := authInterceptor.accessibleRoles[fullMethod]
	if !exists {
		// If method is not in the accessibleRoles map, assume it's public
		return nil
	}

	authHeader := []string{r.Header.Get("Authorization")}
	return authInterceptor.verifyAuth(authHeader, allowedRoles)
}

func (authInterceptor *AuthInterceptor) verifyAuth(authHeader []string, allowedRoles []string) error {
	if len(authHeader) < 1 {
		return status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	fields := strings.Fields(authHeader[0])
	if len(fields) < 2 {
		return status.Errorf(codes.Unauthenticated, "invalid token format")
	}

	if strings.ToLower(fields[0]) != middlewares.AuthorizationType {
		return status.Errorf(codes.Unauthenticated, "unsupported authorization type %v", middlewares.AuthorizationType)
	}

	token := fields[1]
	payload, err := authInterceptor.tokenMaker.VerifyToken(token)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "token verification failed")
	}

	for _, role := range allowedRoles {
		if role == string(payload.Role) {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "user role '%s' does not have access to this resource", payload.Role)
}
