package postgres

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

func TestGormUserRepository(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&UserModel{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	repo, err := NewGormUserRepository(db)
	if err != nil {
		t.Fatalf("repo init: %v", err)
	}

	u, err := repo.Create(&user.User{Email: "user@example.com", PasswordHash: mustHash(t, "secret"), Roles: []string{"user"}, Permissions: []string{"*"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.Email != "user@example.com" {
		t.Fatalf("email mismatch")
	}

	if _, err := repo.Create(&user.User{Email: "user@example.com", PasswordHash: mustHash(t, "secret")}); err == nil {
		t.Fatalf("expected duplicate error")
	}

	got, err := repo.FindByEmail("user@example.com")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("id mismatch")
	}

	if _, err := repo.FindByEmail("missing@example.com"); err == nil {
		t.Fatalf("expected not found")
	}

	found, err := repo.FindByID(u.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if found.Email != u.Email {
		t.Fatalf("find by id mismatch")
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := user.HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return hash
}
