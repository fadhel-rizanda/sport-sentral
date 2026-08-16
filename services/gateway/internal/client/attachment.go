package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	attachmentv1 "microservice-golang/gen/attachment/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type AttachmentClient struct {
	Attachment attachmentv1.AttachmentServiceClient
	conn       *grpc.ClientConn
}

func NewAttachmentClient(address string) (*AttachmentClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &AttachmentClient{
		Attachment: attachmentv1.NewAttachmentServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *AttachmentClient) Close() error {
	return c.conn.Close()
}
