package gapi

import (
	"context"
	"main/pkg/interceptors"
	"main/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetAuthPayload(ctx context.Context) (*token.Payload, error) {
	payload, ok := ctx.Value(interceptors.AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return nil, status.Errorf(codes.Internal, "auth payload not found in context")
	}
	return payload, nil
}
