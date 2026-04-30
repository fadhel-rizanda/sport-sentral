package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	authv1 "microservice-golang/gen/auth/v1"
	rbacv1 "microservice-golang/gen/rbac/v1"
	userv1 "microservice-golang/gen/user/v1"
)

type IdentityClient struct {
	Auth authv1.AuthServiceClient
	User userv1.UserServiceClient
	RBAC rbacv1.RbacServiceClient
	conn *grpc.ClientConn
}

func NewIdentityClient(address string) (*IdentityClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &IdentityClient{
		Auth: authv1.NewAuthServiceClient(conn),
		User: userv1.NewUserServiceClient(conn),
		RBAC: rbacv1.NewRbacServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *IdentityClient) Close() error {
	return c.conn.Close()
}
