package interceptors

import (
	"context"
	"main/pkg/log"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (authInterceptor *AuthInterceptor) LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		startTime := time.Now()
		result, err := handler(ctx, req)

		duration := time.Since(startTime)

		statusCode := codes.Unknown

		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		}
		fields := logrus.Fields{
			"duration":    duration,
			"status":      int(statusCode),
			"status_text": statusCode,
			"method":      info.FullMethod,
		}
		if err != nil {
			log.Logger.WithFields(fields).Error("gRPC request failed")
		} else {
			log.Logger.WithFields(fields).Info("gRPC request processed")
		}
		return result, err
	}

}
