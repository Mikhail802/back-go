package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID    `gorm:"primaryKey" json:"id"`
	OwnerID   uuid.UUID    `json:"owner_id"`
	Name      string       `json:"name"`
	Theme     string       `json:"theme"`
	CreatedAt time.Time    `json:"created_at"`


	Members []RoomMember `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"members"`

}
