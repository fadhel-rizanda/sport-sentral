package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	scoutv1 "microservice-golang/gen/scout/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type ScoutClient struct {
	Scout scoutv1.ScoutServiceClient
	conn  *grpc.ClientConn
}

func NewScoutClient(address string) (*ScoutClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &ScoutClient{
		Scout: scoutv1.NewScoutServiceClient(conn),
		conn:  conn,
	}, nil
}

func (c *ScoutClient) Close() error {
	return c.conn.Close()
}
