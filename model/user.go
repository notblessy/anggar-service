package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	Authenticate(ctx context.Context, code, requestOrigin string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
}

type User struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Password  string         `json:"password,omitempty"`
	Picture   string         `json:"picture"`
	Role      string         `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (u *User) OmitPassword() {
	u.Password = ""
}

type Auth struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type AuthRequest struct {
	Code          string `json:"code"`
	RequestOrigin string `json:"request_origin"`
}

type ChangeUsernameRequest struct {
	Username string `json:"username"`
}

type GoogleAuthInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}
