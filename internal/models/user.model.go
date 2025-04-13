package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет модель пользователя
type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`
	Name              string    `json:"name"`
	Username          string    `gorm:"unique;not null"`
	Email             string    `gorm:"unique" json:"email"`
	Picture  		  string
	Password          string    `json:"-"`
	PasswordResetCode string    `json:"-"`
	CodeExpiry        time.Time `json:"-"`
	CodeUsed          bool      `json:"-"`
}

// BeforeCreate устанавливает UUID перед созданием записи
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.ID = uuid.New()
	return
}

// CreateUser Schema представляет данные для создания нового пользователя
type CreateUserSchema struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUser Schema представляет данные для обновления пользователя
type UpdateUserSchema struct {
	Name     string `json:"name" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// LoginUser Schema представляет данные для авторизации пользователя
type LoginUserSchema struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
