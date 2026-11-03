package test

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/doutorfinancas/pun-sho/database"
)

func NewMockDB() (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("can't create sqlmock: %s", err)
	}

	return mock, openGorm(db)
}

// NewMockDBWithConverter is NewMockDB for models whose Valuer the default sqlmock converter rejects.
func NewMockDBWithConverter(c driver.ValueConverter) (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(c))
	if err != nil {
		log.Fatalf("can't create sqlmock: %s", err)
	}

	return mock, openGorm(db)
}

func openGorm(db *sql.DB) *gorm.DB {
	d := postgres.New(
		postgres.Config{
			Conn:       db,
			DriverName: "postgresql",
		},
	)

	gormDB, gerr := gorm.Open(d, &gorm.Config{})
	if gerr != nil {
		log.Fatalf("can't open gorm connection: %s", gerr)
	}

	return gormDB.Set("gorm:update_column", true)
}

func NewMockRepository(db *gorm.DB) database.Repository {
	return database.Repository{
		Database: database.NewDatabase(db),
	}
}

func CheckMockDB(t *testing.T, mock sqlmock.Sqlmock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func GenerateMockRows(headers []string, rows [][]driver.Value) *sqlmock.Rows {
	r := sqlmock.NewRows(headers)
	for _, w := range rows {
		r = r.AddRow(w...)
	}
	return r
}
