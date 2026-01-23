package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Testzyler/go-microservice/global/pkg/auth"
	authrepo "github.com/Testzyler/go-microservice/services/auth/internal/adapters/postgres"
	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

type stubTokenService struct {
	issued []auth.Claims
	claims auth.Claims
	err    error
}

func (s *stubTokenService) IssueAccessToken(_ context.Context, claims auth.Claims) (string, error) {
	s.issued = append(s.issued, claims)
	if s.err != nil {
		return "", s.err
	}
	return "token-" + claims.Subject.String(), nil
}

func (s *stubTokenService) ParseAccessToken(_ context.Context, tokenStr string) (auth.Claims, error) {
	if s.err != nil {
		return auth.Claims{}, s.err
	}
	return s.claims, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newRepo(t)
	tokens := &stubTokenService{}
	svc := NewAuthService(repo, tokens, defaultTTL)

	res, err := svc.Register(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if res.UserID == uuid.Nil || res.AccessToken == "" {
		t.Fatalf("missing registration outputs")
	}
	if len(tokens.issued) != 1 {
		t.Fatalf("expected token issued on register")
	}

	login, err := svc.Login(context.Background(), "user@example.com", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.AccessToken == "" {
		t.Fatalf("missing login token")
	}
	if len(tokens.issued) != 2 {
		t.Fatalf("expected token issued on login")
	}
}

func TestAuthService_LoginInvalid(t *testing.T) {
	repo := newRepo(t)
	tokens := &stubTokenService{}
	svc := NewAuthService(repo, tokens, defaultTTL)

	_, _ = svc.Register(context.Background(), "user@example.com", "secret")
	if _, err := svc.Login(context.Background(), "user@example.com", "wrong"); !errors.Is(err, user.ErrBadPassword) {
		t.Fatalf("expected ErrBadPassword, got %v", err)
	}
}

func TestAuthService_Validate(t *testing.T) {
	repo := newRepo(t)
	tokens := &stubTokenService{
		claims: auth.Claims{
			Subject: uuid.New(),
			Roles:   []string{"user"},
		},
	}
	svc := NewAuthService(repo, tokens, defaultTTL)

	claims, err := svc.Validate(context.Background(), "any")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if claims.Subject == uuid.Nil {
		t.Fatalf("expected subject")
	}
}

var defaultTTL = 15 * time.Minute

func newRepo(t *testing.T) authrepo.UserRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&authrepo.UserModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo, err := authrepo.NewGormUserRepository(db)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return repo
}
