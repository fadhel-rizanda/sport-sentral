package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	authv1 "microservice-golang/gen/auth/v1"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
)

type IdentityClient struct {
	Auth       authv1.AuthServiceClient
	User       userv1.UserServiceClient
	RBAC       rbacv1.RBACServiceClient
	Role       rbacv1.RoleServiceClient
	Permission rbacv1.PermissionServiceClient
	conn       *grpc.ClientConn
}

func NewIdentityClient(address string) (*IdentityClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &IdentityClient{
		Auth:       authv1.NewAuthServiceClient(conn),
		User:       userv1.NewUserServiceClient(conn),
		RBAC:       rbacv1.NewRBACServiceClient(conn),
		Role:       rbacv1.NewRoleServiceClient(conn),
		Permission: rbacv1.NewPermissionServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *IdentityClient) Close() error {
	return c.conn.Close()
}
