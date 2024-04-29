package rpc_client

import intrvproto "github.com/RafalSalwa/auth-api/proto/grpc"

func NewUserClient(addr string) (intrvproto.UserServiceClient, error) {
	conn, err := newConnection(addr)
	if err != nil {
		return nil, err
	}
	return intrvproto.NewUserServiceClient(conn), nil
}
