package command

import (
	"context"

	"github.com/RafalSalwa/auth-api/pkg/hashing"
	intrvproto "github.com/RafalSalwa/auth-api/proto/grpc"
)

type (
	ChangePassword struct {
		Id       int64
		Password string
	}
	ChangePasswordHandler struct {
		grpcUser intrvproto.UserServiceClient
	}
)

func NewChangePasswordHandler(grpcUser intrvproto.UserServiceClient) ChangePasswordHandler {
	return ChangePasswordHandler{grpcUser: grpcUser}
}

func (h ChangePasswordHandler) Handle(ctx context.Context, cmd ChangePassword) error {
	passHash, err := hashing.HashPassword(cmd.Password)
	if err != nil {
		return err
	}
	_, err = h.grpcUser.ChangePassword(ctx, &intrvproto.ChangePasswordRequest{
		Id:       cmd.Id,
		Password: passHash,
	})
	if err != nil {
		return err
	}
	return nil
}
