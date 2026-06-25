package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	academyv1 "microservice-golang/gen/academy/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type AcademyClient struct {
	AcademyHolding academyv1.AcademyHoldingServiceClient
	AcademyBranch  academyv1.AcademyBranchServiceClient
	AcademyAdmin   academyv1.AcademyAdminServiceClient
	Enrollment     academyv1.EnrollmentServiceClient
	conn           *grpc.ClientConn
}

func NewAcademyClient(address string) (*AcademyClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &AcademyClient{
		AcademyHolding: academyv1.NewAcademyHoldingServiceClient(conn),
		AcademyBranch:  academyv1.NewAcademyBranchServiceClient(conn),
		AcademyAdmin:   academyv1.NewAcademyAdminServiceClient(conn),
		Enrollment:     academyv1.NewEnrollmentServiceClient(conn),
		conn:           conn,
	}, nil
}

func (c *AcademyClient) Close() error {
	return c.conn.Close()
}
