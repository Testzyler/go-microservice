package user

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")

	ErrNotFound     = errors.New("user not found")
	ErrBadPassword  = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
)
