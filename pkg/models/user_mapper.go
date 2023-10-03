package models

import (
	"fmt"
	"github.com/RafalSalwa/interview-app-srv/pkg/encdec"
	intrvproto "github.com/RafalSalwa/interview-app-srv/proto/grpc"

	"github.com/RafalSalwa/interview-app-srv/pkg/jwt"
	"github.com/jinzhu/copier"
)

func (r *UserResponse) FromDBResponse(user *UserDBResponse) error {
	err := copier.Copy(r, &user)
	if err != nil {
		return fmt.Errorf("from db to response error: %w", err)
	}
	return nil
}

func (m *UserDBModel) FromCreateUserReq(cur *SignUpUserRequest) error {
	err := copier.Copy(m, &cur)
	m.Username = cur.Email
	if err != nil {
		return fmt.Errorf("from create to db model error: %w", err)
	}

	return nil
}

func (r *UserResponse) FromProtoUserDetails(pbu *intrvproto.UserDetails) {
	r.Id = pbu.GetId()
	r.Username = pbu.GetUsername()
	r.Firstname = pbu.GetFirstname()
	r.Lastname = pbu.GetLastname()
	r.Verified = pbu.GetVerified()
	r.Active = pbu.GetActive()
	r.CreatedAt = pbu.GetCreatedAt().AsTime()

	ll := pbu.GetLastLogin().AsTime()
	r.LastLogin = &ll
	if pbu.GetEmail() != "" {
		dec, _ := encdec.Decrypt(pbu.GetEmail())
		r.Email = dec
	}
}

func (r *UserResponse) FromDBModel(um *UserDBModel) error {
	err := copier.Copy(r, &um)
	if err != nil {
		return fmt.Errorf("from response to db error: %w", err)
	}
	r.Username = um.Username
	r.CreatedAt = um.CreatedAt
	return nil
}

func (r *UserResponse) AssignTokenPair(tp *jwt.TokenPair) {
	r.AccessToken = tp.AccessToken
	r.RefreshToken = tp.RefreshToken
}

func (r *UserResponse) FromProtoSignIn(pbu *intrvproto.SignInUserResponse) {
	r.AccessToken = pbu.AccessToken
	r.RefreshToken = pbu.RefreshToken
}
func (r *UserResponse) FromProtoSignUp(pbu *intrvproto.SignUpUserResponse) error {
	r.Username = pbu.GetUsername()
	r.VerificationCode = pbu.GetVerificationToken()
	c := pbu.GetCreatedAt().AsTime()
	r.CreatedAt = c
	return nil
}

func (r *UserResponse) FromProtoUserResponse(pu *intrvproto.UserResponse) error {
	err := copier.Copy(r, &pu)
	r.Id = pu.User.Id
	r.Username = pu.User.Username

	if err != nil {
		return fmt.Errorf("from response to db error: %w", err)
	}
	return nil
}

func (r *UserDBResponse) FromProtoUserDetails(pw *intrvproto.UserDetails) error {
	err := copier.Copy(r, pw)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserMongoModel) FromDBModel(user *UserDBModel) error {
	err := copier.Copy(m, user)
	if err != nil {
		return err
	}
	return nil
}
