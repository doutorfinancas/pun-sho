package database

import (
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/model"
)

type Repository struct {
	Database *Database
	Logger   *zap.Logger
}

func (r *Repository) Create(m model.Model) error {
	return r.Database.Create(m)
}

func (r *Repository) Find(m model.Model) error {
	return r.Database.FetchOne(m)
}

func (r *Repository) Save(m model.Model) error {
	return r.Database.Save(m)
}
