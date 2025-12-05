package database

import (
	"razpravljalnica/internal/api"
)

func ToApiUser(user *User) *api.User {
	return &api.User{
		Id:   user.ID,
		Name: user.Name,
	}
}
