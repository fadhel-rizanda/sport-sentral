package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	competitionv1 "microservice-golang/gen/competition/v1"
	"microservice-golang/shared/pkg/grpc/interceptor"
)

type CompetitionClient struct {
	Roster      competitionv1.RosterServiceClient
	Competition competitionv1.CompetitionServiceClient
	Branch      competitionv1.CompetitionBranchServiceClient
	Match       competitionv1.MatchServiceClient
	Stat        competitionv1.StatServiceClient
	conn        *grpc.ClientConn
}

func NewCompetitionClient(address string) (*CompetitionClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientMetadataPropagator()),
	)
	if err != nil {
		return nil, err
	}

	return &CompetitionClient{
		Roster:      competitionv1.NewRosterServiceClient(conn),
		Competition: competitionv1.NewCompetitionServiceClient(conn),
		Branch:      competitionv1.NewCompetitionBranchServiceClient(conn),
		Match:       competitionv1.NewMatchServiceClient(conn),
		Stat:        competitionv1.NewStatServiceClient(conn),
		conn:        conn,
	}, nil
}

func (c *CompetitionClient) Close() error {
	return c.conn.Close()
}
