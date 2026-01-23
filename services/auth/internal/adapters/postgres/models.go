package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

type UserModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email        string         `gorm:"uniqueIndex;not null"`
	PasswordHash string         `gorm:"not null"`
	Roles        pq.StringArray `gorm:"type:text[]"`
	Permissions  pq.StringArray `gorm:"type:text[]"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string {
	return "auth_users"
}

func userModelFromDomain(u *user.User) UserModel {
	return UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Roles:        toStringArray(u.Roles),
		Permissions:  toStringArray(u.Permissions),
		CreatedAt:    u.CreatedAt,
	}
}

func (m UserModel) ToDomain() *user.User {
	return &user.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Roles:        append([]string(nil), m.Roles...),
		Permissions:  append([]string(nil), m.Permissions...),
		CreatedAt:    m.CreatedAt,
	}
}

func toStringArray(values []string) pq.StringArray {
	if len(values) == 0 {
		return pq.StringArray{}
	}
	return pq.StringArray(values)
}
