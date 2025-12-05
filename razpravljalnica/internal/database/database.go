package database

import (
	"context"
	"errors"
	"fmt"
	"razpravljalnica/internal/shared"
	"sync/atomic"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db         *gorm.DB
	version    atomic.Uint64
	observable *shared.Observable[uint64]
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
		db:         db,
		observable: shared.NewObservable[uint64](),
	}, nil
}

func (d *Database) ObserveVersion(ctx context.Context, observerID string) (<-chan *uint64, func()) {
	return d.observable.Observe(ctx, observerID, 0)
}

func (d *Database) Save(ctx context.Context, value any) error {
	result := d.db.Save(value)
	if result.Error != nil {
		return result.Error
	}

	newVersion := d.version.Add(1)
	d.observable.Notify(ctx, 0, &newVersion)

	return nil
}

func (d *Database) Delete(ctx context.Context, value any, conds ...any) error {
	result := d.db.Delete(value, conds...)
	if result.Error != nil {
		return result.Error
	}

	newVersion := d.version.Add(1)
	d.observable.Notify(ctx, 0, &newVersion)

	return nil
}

func (d *Database) FindUserByID(ctx context.Context, id int64) (user *User, err error) {
	result := d.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return user, nil
}

func (d *Database) FindTopicByID(ctx context.Context, id int64) (topic *Topic, err error) {
	result := d.db.First(&topic, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return topic, nil
}
