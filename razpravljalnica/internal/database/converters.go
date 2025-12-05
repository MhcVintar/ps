package database

import (
	"razpravljalnica/internal/api"
)

func ToApiWALEntry(walEntry *WALEntry) *api.WALEntry {
	return &api.WALEntry{
		Id:     walEntry.ID,
		Op:     walEntry.Op,
		Target: walEntry.Target,
		Data:   walEntry.Data,
	}
}

func ToApiUser(user *User) *api.User {
	return &api.User{
		Id:   user.ID,
		Name: user.Name,
	}
}
