package entity

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type ShortyRepository struct {
	database.Repository
}

func NewShortyRepository(db *database.Database, log *zap.Logger) *ShortyRepository {
	return &ShortyRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

func (r *ShortyRepository) List(limit, offset int) ([]*Shorty, error) {
	rows := make([]*Shorty, 0)

	if err := r.Database.FetchPage(
		&Shorty{},
		limit,
		offset,
		&rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *ShortyRepository) Delete(id uuid.UUID) error {
	m := Shorty{ID: id}
	return r.Database.Orm.Model(m).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).
		Error
}
