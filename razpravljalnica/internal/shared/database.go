package shared

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.NewSlogLogger(Logger, logger.Config{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&User{},
		&Topic{},
		&Message{},
		&Like{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	Logger.Info("connected to database")

	return db, nil
}

type User struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);not null" json:"name"`
}

func (*User) TableName() string {
	return "users"
}

type Topic struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);not null" json:"name"`
}

func (*Topic) TableName() string {
	return "topics"
}

type Message struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TopicID   int64     `gorm:"not null;index" json:"topic_id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Text      string    `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	Likes     int32     `gorm:"default:0" json:"likes"`
}

func (*Message) TableName() string {
	return "messages"
}

type Like struct {
	TopicID   int64 `gorm:"primaryKey;not null;index:idx_like" json:"topic_id"`
	MessageID int64 `gorm:"primaryKey;not null;index:idx_like" json:"message_id"`
	UserID    int64 `gorm:"primaryKey;not null;index:idx_like" json:"user_id"`
}

func (*Like) TableName() string {
	return "likes"
}
