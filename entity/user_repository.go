package entity

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type UserRepository struct {
	database.Repository
}

func NewUserRepository(db *database.Database, log *zap.Logger) *UserRepository {
	return &UserRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

func (r *UserRepository) FindByUsername(username string) (*User, error) {
	user := &User{}
	if err := r.Database.Orm.
		Where("username = ? AND deleted_at IS NULL", username).
		First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*User, error) {
	user := &User{}
	if err := r.Database.Orm.
		Where("email = ? AND deleted_at IS NULL", email).
		First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByMSEmail(msEmail string) (*User, error) {
	user := &User{}
	if err := r.Database.Orm.
		Where("ms_email = ? AND deleted_at IS NULL", msEmail).
		First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id uuid.UUID) (*User, error) {
	user := &User{}
	if err := r.Database.Orm.
		Where("id = ? AND deleted_at IS NULL", id).
		First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) List() ([]*User, error) {
	var users []*User
	if err := r.Database.Orm.
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) HardDelete(id uuid.UUID) error {
	return r.Database.Orm.
		Where("id = ?", id).
		Delete(&User{}).Error
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.Database.Orm.
		Model(&User{}).
		Where("deleted_at IS NULL").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
