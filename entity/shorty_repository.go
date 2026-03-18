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

func (r *ShortyRepository) ListWithAccessData(listWithQRCode bool, labels []string, limit, offset int) ([]*ShortyForList, error) {
	rows := make([]*ShortyForList, 0)

	fields := []string{
		"s.id",
		"s.created_at",
		"s.deleted_at",
		"s.public_id",
		"s.link",
		"s.ttl",
		"s.redirection_limit",
		"s.labels",
	}

	if listWithQRCode {
		fields = append(fields, "qr_code")
	}

	fieldlist := strings.Join(fields, ", ")

	baseQuery := fmt.Sprintf(
		`SELECT %s, count(sa.id) as visits, COALESCE(sum(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END), 0) as redirects FROM shorties s
    LEFT JOIN shorty_accesses sa
        ON s.id = sa.shorty_id`, fieldlist)

	var whereClause string
	var args []interface{}

	if len(labels) > 0 {
		// Filter by labels: check if any of the provided labels exist in the labels array
		whereClause = " WHERE s.labels && ?"
		args = append(args, fmt.Sprintf("{%s}", strings.Join(labels, ",")))
	}

	groupByClause := " GROUP BY s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.qr_code, s.ttl, s.redirection_limit, s.labels LIMIT ? OFFSET ?"

	query := baseQuery + whereClause + groupByClause
	args = append(args, limit, offset)

	if err := r.Database.Orm.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *ShortyRepository) ExistsByPublicID(publicID string) (bool, error) {
	var count int64
	if err := r.Database.Orm.
		Model(&Shorty{}).
		Where("public_id = ?", publicID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ShortyRepository) Delete(id uuid.UUID) error {
	m := Shorty{ID: id}
	return r.Database.Orm.Model(m).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).
		Error
}
