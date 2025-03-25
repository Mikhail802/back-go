package models

import "github.com/google/uuid"

type Friendship struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid"`
	FriendID uuid.UUID `gorm:"type:uuid"`
	Status   string
	User     User      `gorm:"foreignKey:UserID" json:"user"` // для Preload
}

