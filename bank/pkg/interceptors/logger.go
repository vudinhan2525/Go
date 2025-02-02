package interceptors

import (
	"context"
	"main/pkg/log"
	"net/http"
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

type ResponseRecoder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecoder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
func (rec *ResponseRecoder) Write(body []byte) (int, error) {
	rec.Body = append(rec.Body, body...)
	return rec.ResponseWriter.Write(body)
}

func (authInterceptor *AuthInterceptor) LoggerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		rec := &ResponseRecoder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}

		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)
		fields := logrus.Fields{
			"duration":    duration,
			"status":      int(rec.StatusCode),
			"status_text": http.StatusText(rec.StatusCode),
			"method":      req.Method,
		}
		if rec.StatusCode != http.StatusOK {
			log.Logger.WithFields(fields).Error("http request failed")
		} else {
			log.Logger.WithFields(fields).Info("http request processed")
		}

	})

}
