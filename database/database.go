package database

import (
	"errors"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/doutorfinancas/pun-sho/model"
)

type Database struct {
	Orm *gorm.DB
}

func NewDatabase(orm *gorm.DB) *Database {
	return &Database{
		Orm: orm,
	}
}

func (d *Database) FetchOne(mod model.Model) error {
	return d.Orm.First(mod, mod).Error
}

func (d *Database) FetchAll(mod model.Model, rows interface{}) error {
	return d.Orm.Model(mod).Where(mod).Find(rows).Error
}

func (d *Database) FetchOnly(mod model.Model, limit int, rows interface{}) error {
	return d.Orm.Model(mod).Where(mod).Limit(limit).Find(rows).Error
}

func (d *Database) FetchLatest(mod model.Model, creationColumn string) error {
	return d.Orm.Model(mod).Order(
		clause.OrderByColumn{
			Column: clause.Column{Name: creationColumn},
			Desc:   true,
		},
	).Limit(1).Error
}

func (d *Database) FetchPage(
	mod model.Model,
	limit int,
	offset int,
	rows interface{},
) error {
	return d.Orm.Model(mod).Where(mod).Limit(limit).Offset(offset).Find(rows).Error
}

func (d *Database) CountAll(mod model.Model) int64 {
	var count int64

	d.Orm.Model(mod).Where(mod).Count(&count)

	return count
}

func (d *Database) Create(mod model.Model) error {
	if reflect.ValueOf(mod).Kind() == reflect.Ptr {
		if reflect.ValueOf(mod).IsNil() {
			return errors.New("cannot pass nil model to create something")
		}
	}

	return d.Orm.Create(mod).Error
}

func (d *Database) Save(mod model.Model) error {
	val := reflect.ValueOf(mod)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return errors.New("cannot pass nil model to save something")
		}
		val = val.Elem()
	}

	f := val.FieldByName("ID")
	if reflect.Invalid != f.Kind() && f.String() == "" {
		return errors.New("you must pass a valid non empty ID in order to save it")
	}

	if err := d.Orm.Model(mod).Updates(mod).Error; err != nil {
		return err
	}

	d.Orm.Model(mod).Find(mod)

	return nil
}

func (d *Database) Upsert(mod model.Model) error {
	val := reflect.ValueOf(mod)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return errors.New("cannot pass nil model to save something")
		}
		val = val.Elem()
	}

	f := val.FieldByName("ID")

	if (f.Kind() == reflect.Int && f.Int() == 0) || (f.Kind() == reflect.String && f.String() == "") {
		return d.Create(mod)
	}
	return d.Save(mod)
}

func (d *Database) Transaction(transaction func(tx *gorm.DB) error) error {
	return d.Orm.Transaction(transaction)
}

func (d *Database) Delete(mod model.Model) error {
	return d.Orm.Model(mod).Where(mod).Delete(mod).Error
}
