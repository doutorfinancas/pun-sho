package entity

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type SessionRepository struct {
	database.Repository
}

func NewSessionRepository(db *database.Database, log *zap.Logger) *SessionRepository {
	return &SessionRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

func (r *SessionRepository) FindByToken(token string) (*Session, error) {
	session := &Session{}
	if err := r.Database.Orm.
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) DeleteByToken(token string) error {
	return r.Database.Orm.
		Where("token = ?", token).
		Delete(&Session{}).Error
}

func (r *SessionRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.Database.Orm.
		Where("user_id = ?", userID).
		Delete(&Session{}).Error
}

func (r *SessionRepository) DeleteExpired() error {
	return r.Database.Orm.
		Where("expires_at < ?", time.Now()).
		Delete(&Session{}).Error
}
