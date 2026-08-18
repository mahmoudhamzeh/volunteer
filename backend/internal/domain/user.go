package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleVolunteer Role = "volunteer"
	RoleAdmin     Role = "admin"
	RoleOperator  Role = "operator"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Phone          string
	PasswordHash   string
	Role           Role
	ExternalUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u User) IsStaff() bool {
	return u.Role == RoleAdmin || u.Role == RoleOperator
}
