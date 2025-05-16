package repository

import (
	"time"

	"auto-message-sender/internal/entity"
	"auto-message-sender/internal/model/request"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *entity.Message) error
	GetUnsentMessages(limit int) ([]entity.Message, error)
	UpdateStatus(messageID, status string, sentAt time.Time) error
	UpdateMessageID(id uuid.UUID, messageID string) error
	GetMessages(filter *request.MessageFilterRequest) ([]entity.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *entity.Message) error {
	message.ID = uuid.New()
	return r.db.Create(message).Error
}

func (r *messageRepository) GetUnsentMessages(limit int) ([]entity.Message, error) {
	var messages []entity.Message
	err := r.db.Where("status = ?", entity.StatusPending).Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetMessages(filter *request.MessageFilterRequest) ([]entity.Message, error) {
	var messages []entity.Message
	query := r.db.Model(&entity.Message{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status = ?", entity.StatusSent)
	}

	if filter.StartDate != "" {
		startDate, _ := time.Parse("2006-01-02", filter.StartDate)
		query = query.Where("sent_at >= ?", startDate)
	}

	if filter.EndDate != "" {
		endDate, _ := time.Parse("2006-01-02", filter.EndDate)
		endDatePlusDay := endDate.AddDate(0, 0, 1)
		query = query.Where("sent_at < ?", endDatePlusDay)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	err := query.Order("sent_at DESC").Find(&messages).Error
	return messages, err
}

func (r *messageRepository) UpdateStatus(messageID, status string, sentAt time.Time) error {
	return r.db.Model(&entity.Message{}).
		Where("message_id = ?", messageID).
		Updates(map[string]interface{}{
			"status":  status,
			"sent_at": sentAt,
		}).Error
}

func (r *messageRepository) UpdateMessageID(id uuid.UUID, messageID string) error {
	return r.db.Model(&entity.Message{}).
		Where("id = ?", id).
		Update("message_id", messageID).Error
}
