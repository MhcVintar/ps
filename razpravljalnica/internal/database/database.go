package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db         *gorm.DB
	walQueue   chan *WALEntry
	freezeLock sync.RWMutex
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(sqlite.Open("file:razpravljalnica?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.NewSlogLogger(shared.Logger, logger.Config{}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&WALEntry{},
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

func (d *Database) Freeze() {
	d.freezeLock.Lock()
}

func (d *Database) Unfreeze() {
	d.freezeLock.Unlock()
}

func (d *Database) WALQueue() <-chan *WALEntry {
	return d.walQueue
}

func (d *Database) LSN() (int64, error) {
	var maxID int64
	err := d.db.Model(&WALEntry{}).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID).Error
	if err != nil {
		return 0, err
	}
	return maxID, nil
}

func (d *Database) YieldWAL(ctx context.Context, fromLSN int64) <-chan *WALEntry {
	out := make(chan *WALEntry, 100)

	go func() {
		defer close(out)

		rows, err := d.db.WithContext(ctx).
			Model(&WALEntry{}).
			Where("id >= ?", fromLSN).
			Order("id ASC").
			Rows()
		if err != nil {
			shared.Logger.Error("failed to stream WAL entries", "err", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var entry WALEntry
			if err := d.db.ScanRows(rows, &entry); err != nil {
				shared.Logger.Error("failed to scan WAL row", "err", err)
				return
			}

			select {
			case out <- &entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (d *Database) Save(ctx context.Context, value any) error {
	d.freezeLock.RLock()
	defer d.freezeLock.Unlock()

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
	d.freezeLock.RLock()
	defer d.freezeLock.Unlock()

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
