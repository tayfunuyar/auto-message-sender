package database

import (
	"auto-message-sender/internal/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedMessages(db *gorm.DB) error {
	var count int64
	if err := db.Model(&entity.Message{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	messages := []entity.Message{
		{
			ID:      uuid.New(),
			To:      "+905551111111",
			Content: "Test message 1",
			Status:  entity.StatusPending,
		},
		{
			ID:      uuid.New(),
			To:      "+905552222222",
			Content: "Test message 2",
			Status:  entity.StatusPending,
		},
		{
			ID:      uuid.New(),
			To:      "+905553333333",
			Content: "Test message 3",
			Status:  entity.StatusPending,
		},
		{
			ID:        uuid.New(),
			To:        "+905554444444",
			Content:   "Test message 4",
			Status:    entity.StatusSent,
			MessageID: "test-message-id-1",
			SentAt:    time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New(),
			To:        "+905555555555",
			Content:   "Test message 5",
			Status:    entity.StatusSent,
			MessageID: "test-message-id-2",
			SentAt:    time.Now().Add(-12 * time.Hour),
		},
	}

	return db.CreateInBatches(messages, len(messages)).Error
}
