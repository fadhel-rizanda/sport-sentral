package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	logv1 "microservice-golang/gen/log/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type LogClient struct {
	Log  logv1.LogServiceClient
	conn *grpc.ClientConn
}

func NewLogClient(address string) (*LogClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &LogClient{
		Log:  logv1.NewLogServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *LogClient) Close() error {
	return c.conn.Close()
}
