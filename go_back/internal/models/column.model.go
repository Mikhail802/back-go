package models

import (
	"github.com/google/uuid"
	"time"
)

type Column struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	RoomID    uuid.UUID `gorm:"not null;index" json:"room_id"`
	Title     string    `json:"title"`
	Tasks     []Task    `gorm:"foreignKey:ColumnID;references:ID;constraint:OnDelete:CASCADE" json:"tasks"` // ✅ Правильная связь
	CreatedAt time.Time `json:"created_at"`
}
