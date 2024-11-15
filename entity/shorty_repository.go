package entity

import (
	"fmt"
	"strings"
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

func (r *ShortyRepository) ListWithAccessData(listWithQRCode bool, limit, offset int) ([]*ShortyForList, error) {
	rows := make([]*ShortyForList, 0)

	fields := []string{
		"s.id",
		"s.created_at",
		"s.deleted_at",
		"s.public_id",
		"s.link",
		"s.ttl",
		"s.redirection_limit",
	}

	if listWithQRCode {
		fields = append(fields, "qr_code")
	}

	fieldlist := strings.Join(fields, ", ")

	query := fmt.Sprintf(
		`SELECT %s, count(sa.id) as visits, sum(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END) as redirects FROM shorties s 
    INNER JOIN shorty_accesses sa 
        ON s.id = sa.shorty_id 
    GROUP BY s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.qr_code, s.ttl, s.redirection_limit
LIMIT ? OFFSET ?`, fieldlist)

	if err := r.Database.Orm.Raw(query, limit, offset).Scan(&rows).Error; err != nil {
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
