package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	sportv1 "microservice-golang/gen/sport/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type SportClient struct {
	Sport     sportv1.SportServiceClient
	Regulator sportv1.RegulatorServiceClient
	conn      *grpc.ClientConn
}

func NewSportClient(address string) (*SportClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &SportClient{
		Sport:     sportv1.NewSportServiceClient(conn),
		Regulator: sportv1.NewRegulatorServiceClient(conn),
		conn:      conn,
	}, nil
}

func (c *SportClient) Close() error {
	return c.conn.Close()
}
