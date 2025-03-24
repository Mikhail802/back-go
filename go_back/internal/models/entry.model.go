package models

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	RoomID    uuid.UUID `gorm:"not null" json:"room_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
