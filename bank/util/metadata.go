package util

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	GrpcGatewayAgent = "grpcgateway-user-agent"
	UserAgent        = "user-agent"
	XForwardFor      = "x-forwarded-for"
)

type Metadata struct {
	ClientIp  string
	UserAgent string
}

func ExtractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(GrpcGatewayAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if userIps := md.Get(XForwardFor); len(userIps) > 0 {
			mtdt.ClientIp = userIps[0]
		}

		if userAgents := md.Get(UserAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIp = p.Addr.String()
	}
	return mtdt
}
