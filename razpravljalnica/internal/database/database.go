package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db       *gorm.DB
	walQueue chan *WALEntry
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.NewSlogLogger(shared.Logger, logger.Config{}),
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

	shared.Logger.Info("connected to database")

	return &Database{
		db:       db,
		walQueue: make(chan *WALEntry, 100),
	}, nil
}

func (d *Database) Close() {
	close(d.walQueue)
}

func (d *Database) WALQueue() <-chan *WALEntry {
	return d.walQueue
}

func (d *Database) LSN() (int64, error) {
	var maxID int64
	if err := d.db.Model(&WALEntry{}).Select("MAX(id)").Scan(&maxID).Error; err != nil {
		return 0, err
	}
	return maxID, nil
}

func (d *Database) Save(ctx context.Context, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		walEntry := WALEntry{
			Op:     api.WALEntry_OP_SAVE,
			Target: getTarget(value),
			Data:   data,
		}
		if err := tx.Create(&walEntry).Error; err != nil {
			return err
		}
		d.walQueue <- &walEntry

		if err := tx.Save(value).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *Database) Delete(ctx context.Context, value any, conds ...any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		walEntry := WALEntry{
			Op:     api.WALEntry_OP_DELETE,
			Target: getTarget(value),
			Data:   data,
		}
		if err := tx.Create(&walEntry).Error; err != nil {
			return err
		}
		d.walQueue <- &walEntry

		if err := tx.Delete(value).Error; err != nil {
			return err
		}

		return nil
	})
}

func getTarget(value any) api.WALEntry_Target {
	var target api.WALEntry_Target
	switch value.(type) {
	case *User:
		target = api.WALEntry_TARGET_USER
	case *Topic:
		target = api.WALEntry_TARGET_TOPIC
	case *Message:
		target = api.WALEntry_TARGET_MESSAGE
	case *Like:
		target = api.WALEntry_TARGET_LIKE
	}

	return target
}

func (d *Database) FindUserByID(ctx context.Context, id int64) (user *User, err error) {
	if err := d.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (d *Database) FindTopicByID(ctx context.Context, id int64) (topic *Topic, err error) {
	if err := d.db.First(&topic, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return topic, nil
}
