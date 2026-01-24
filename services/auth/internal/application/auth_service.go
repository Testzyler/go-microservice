package application

import (
	"context"
	"time"

	"github.com/Testzyler/go-microservice/services/auth/internal/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Testzyler/go-microservice/pkg/auth"
	"github.com/Testzyler/go-microservice/pkg/errx"
	"github.com/Testzyler/go-microservice/pkg/token"
	authrepo "github.com/Testzyler/go-microservice/services/auth/internal/adapters/postgres"
	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

type AuthService struct {
	users     authrepo.UserRepository
	tokens    token.Service
	accessTTL time.Duration
	tasks     ports.TaskPublisher
	logger    *zap.Logger
}

func NewAuthService(users authrepo.UserRepository, tokens token.Service, accessTTL time.Duration, tasks ports.TaskPublisher, logger *zap.Logger) *AuthService {
	return &AuthService{
		users:     users,
		tokens:    tokens,
		accessTTL: accessTTL,
		tasks:     tasks,
		logger:    logger,
	}
}

type AuthResult struct {
	UserID      uuid.UUID
	AccessToken string
	Roles       []string
	Permissions []string
	ExpiresIn   int64
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	u, err := user.New(email, password)
	if err != nil {
		return nil, err
	}
	if len(u.Roles) == 0 {
		u.Roles = []string{"user"}
	}
	if len(u.Permissions) == 0 {
		u.Permissions = []string{"*"}
	}
	created, err := s.users.Create(u)
	if err != nil {
		return nil, err
	}
	u = created
	s.enqueueWelcomeEmail(ctx, u)
	return s.issueToken(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	if email == "" || password == "" {
		return nil, errx.New("auth.missing_credentials", "email and password are required", errx.KindInvalidArgument)
	}
	normalized, err := user.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	u, err := s.users.FindByEmail(normalized)
	if err != nil {
		return nil, err
	}
	ok, err := u.CheckPassword(password)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, user.ErrBadPassword
	}
	return s.issueToken(ctx, u)
}

func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.users.FindByID(id)
}

func (s *AuthService) Validate(ctx context.Context, tokenStr string) (*auth.Claims, error) {
	if tokenStr == "" {
		return nil, errx.New("auth.missing_token", "token is required", errx.KindInvalidArgument)
	}
	claims, err := s.tokens.ParseAccessToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func (s *AuthService) issueToken(ctx context.Context, u *user.User) (*AuthResult, error) {
	claims := auth.Claims{
		Subject:     u.ID,
		Roles:       cloneStrings(u.Roles),
		Permissions: cloneStrings(u.Permissions),
	}
	tokenStr, err := s.tokens.IssueAccessToken(ctx, claims)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		UserID:      u.ID,
		AccessToken: tokenStr,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		ExpiresIn:   int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) enqueueWelcomeEmail(ctx context.Context, u *user.User) {
	if s.tasks == nil {
		return
	}
	if err := s.tasks.EnqueueWelcomeEmail(ctx, u.ID.String(), u.Email); err != nil && s.logger != nil {
		s.logger.Warn("enqueue welcome email failed", zap.String("user_id", u.ID.String()), zap.String("email", u.Email), zap.Error(err))
	}
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}
