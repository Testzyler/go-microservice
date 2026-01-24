package user

import "github.com/Testzyler/go-microservice/pkg/errx"

var (
	ErrInvalidEmail       = errx.New("auth.invalid_email", "invalid email", errx.KindInvalidArgument)
	ErrInvalidCredentials = errx.New("auth.invalid_credentials", "invalid credentials", errx.KindUnauthenticated)
	ErrUserExists         = errx.New("auth.user_exists", "user already exists", errx.KindConflict)

	ErrNotFound    = errx.New("auth.user_not_found", "user not found", errx.KindNotFound)
	ErrBadPassword = errx.New("auth.bad_password", "invalid credentials", errx.KindUnauthenticated)
)
