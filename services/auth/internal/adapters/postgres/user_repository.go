package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

type UserRepository interface {
	Create(u *user.User) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
	FindByID(id uuid.UUID) (*user.User, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) (*GormUserRepository, error) {
	return &GormUserRepository{db: db}, nil
}

func (r *GormUserRepository) Create(u *user.User) (*user.User, error) {
	model := &UserModel{
		ID:           uuid.New(),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Roles:        pq.StringArray(append([]string(nil), u.Roles...)),
		Permissions:  pq.StringArray(append([]string(nil), u.Permissions...)),
		CreatedAt:    time.Now().UTC(),
	}

	if err := r.db.Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, user.ErrUserExists
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *GormUserRepository) FindByEmail(email string) (*user.User, error) {
	var model UserModel
	if err := r.db.Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *GormUserRepository) FindByID(id uuid.UUID) (*user.User, error) {
	var model UserModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return errors.Is(err, gorm.ErrDuplicatedKey)
}

var _ UserRepository = (*GormUserRepository)(nil)
