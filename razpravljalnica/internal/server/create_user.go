package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
)

func (s *Server) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*api.User, error) {
	user := shared.User{
		Name: request.Name,
	}
	s.db.Create(&user)
	shared.Logger.InfoContext(ctx, "user created", "user", user)

	return &api.User{
		Id:   user.ID,
		Name: user.Name,
	}, nil
}
