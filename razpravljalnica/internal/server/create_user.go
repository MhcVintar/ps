package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerNode) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.User, error) {
	if !s.isTail() {
		return s.upstreamClient.Public.CreateUser(ctx, req)
	}

	userModel := database.User{
		Name: req.Name,
	}
	if err := s.db.Save(ctx, &userModel); err != nil {
		shared.Logger.ErrorContext(ctx, "failed to save user", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to save user: %v", err)
	}

	shared.Logger.InfoContext(ctx, "created user", "user", userModel)

	return database.ToApiUser(&userModel), nil
}
