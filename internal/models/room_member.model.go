package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomMember struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	RoomID    uuid.UUID `gorm:"type:uuid;not null" json:"room_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Role      string    `gorm:"default:member" json:"role"`
	CreatedAt time.Time `json:"created_at"`


	User User `gorm:"foreignKey:UserID" json:"user"`
}
