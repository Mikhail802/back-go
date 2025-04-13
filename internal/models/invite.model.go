package models

import (
    "time"

    "github.com/google/uuid"
)

type RoomInvite struct {
    ID         uuid.UUID `gorm:"primaryKey"`
    RoomID     uuid.UUID
    ToUserID   uuid.UUID
    FromUserID uuid.UUID
    Status     string    `gorm:"default:pending"` // pending, accepted, rejected
    CreatedAt  time.Time

    Room   Room `gorm:"foreignKey:RoomID"`
    From   User `gorm:"foreignKey:FromUserID"`
    To     User `gorm:"foreignKey:ToUserID"`
}