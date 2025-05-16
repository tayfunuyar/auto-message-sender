package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID        uuid.UUID      `gorm:"type:uuid;primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	To        string         `gorm:"not null" json:"to"`
	Content   string         `gorm:"not null;type:varchar(160)" json:"content"`
	Status    string         `gorm:"not null;default:'pending'" json:"status"`
	MessageID string         `gorm:"index" json:"message_id,omitempty"`
	SentAt    time.Time      `json:"sent_at,omitempty"`
}
