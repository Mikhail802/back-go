package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskAssignment struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE" json:"task_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
	
}
