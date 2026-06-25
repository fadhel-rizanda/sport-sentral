package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type MetaClient struct {
	Status metav1.StatusServiceClient
	Tag    metav1.TagServiceClient
	conn   *grpc.ClientConn
}

func NewMetaClient(address string) (*MetaClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &MetaClient{
		Status: metav1.NewStatusServiceClient(conn),
		Tag:    metav1.NewTagServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *MetaClient) Close() error {
	return c.conn.Close()
}
