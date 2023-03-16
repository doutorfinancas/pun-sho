package entity

import (
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
)

type ShortyAccessRepository struct {
	database.Repository
}

func NewShortyAccessRepository(db *database.Database, log *zap.Logger) *ShortyAccessRepository {
	return &ShortyAccessRepository{
		database.Repository{
			Database: db,
			Logger:   log,
		},
	}
}

func (r *ShortyAccessRepository) ListByShortyUUID(id uuid.UUID, limit, offset int) ([]*ShortyAccess, error) {
	rows := make([]*ShortyAccess, 0)

	if err := r.Database.FetchPage(
		&ShortyAccess{
			ShortyID: id,
		},
		limit,
		offset,
		&rows,
	); err != nil {
		return nil, err
	}

	return rows, nil
}
