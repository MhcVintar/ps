package database

import (
	"razpravljalnica/internal/api"

	"google.golang.org/protobuf/types/known/timestamppb"
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

func ToApiTopic(topic *Topic) *api.Topic {
	return &api.Topic{
		Id:   topic.ID,
		Name: topic.Name,
	}
}

func ToApiMessage(message *Message) *api.Message {
	return &api.Message{
		Id:        message.ID,
		TopicId:   message.TopicID,
		UserId:    message.UserID,
		Text:      message.Text,
		CreatedAt: timestamppb.New(message.CreatedAt),
		Likes:     message.Likes,
	}
}
