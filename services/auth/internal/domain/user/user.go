package user

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Roles        []string
	Permissions  []string
	CreatedAt    time.Time
}

func New(email, password string) (*User, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		Email:        normalized,
		PasswordHash: passwordHash,
		Roles:        []string{},
		Permissions:  []string{},
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func NormalizeEmail(email string) (string, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", ErrInvalidEmail
	}
	return strings.ToLower(address.Address), nil
}

func (u *User) CheckPassword(password string) (bool, error) {
	return VerifyPassword(password, u.PasswordHash)
}

func ValidatePassword(password string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 6 {
		return ErrInvalidPassword
	}
	return nil
}
