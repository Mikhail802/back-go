package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Task struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	RoomID    uuid.UUID `gorm:"type:uuid;not null;index" json:"room_id"`
	ColumnID  uuid.UUID `gorm:"not null;index" json:"column_id"`
	Text      string    `json:"text"`
	Description string  `json:"description"`      // 🔹 добавим описание
	StartDate   *string `json:"startDate"`        // 🔹 даты — в строковом ISO формате
	EndDate     *string `json:"endDate"`
	TaskLists datatypes.JSON `json:"taskLists"`    // 🔥 храним структуру чеклистов как JSON
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`

	AssignedUsers []TaskAssignment `json:"assigned_users"`
}

type TaskListData struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Items []TaskListItem `json:"items"`
}

type TaskListItem struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

