package models

import (
	"github.com/google/uuid"
	"time"
)

type Task struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ColumnID  uuid.UUID `gorm:"not null;index" json:"column_id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}
