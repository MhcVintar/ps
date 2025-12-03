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

	if s.nextServer != nil {
		shared.Logger.InfoContext(ctx, "forwarding request to next server in chain", "request", request)
		apiUser, err := s.nextServer.CreateUser(ctx, request)
		if err != nil {
			return nil, err
		}

		user.ID = apiUser.Id
	}

	s.db.Save(&user)
	shared.Logger.InfoContext(ctx, "user created", "user", user)

	return &api.User{
		Id:   user.ID,
		Name: user.Name,
	}, nil
}
