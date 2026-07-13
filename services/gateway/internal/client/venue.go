package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	venuev1 "microservice-golang/gen/venue/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type VenueClient struct {
	Venue venuev1.VenueServiceClient
	conn  *grpc.ClientConn
}

func NewVenueClient(address string) (*VenueClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &VenueClient{
		Venue: venuev1.NewVenueServiceClient(conn),
		conn:  conn,
	}, nil
}

func (c *VenueClient) Close() error {
	return c.conn.Close()
}
