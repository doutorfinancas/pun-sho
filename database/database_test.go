package database

import (
	"database/sql/driver"
	"log"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/doutorfinancas/pun-sho/test"
)

func getTestModelSingle() *test.TestsModel {
	return &test.TestsModel{
		ID:         999,
		TestInt:    1234,
		TestString: "Test my string",
		TestBool:   true,
		TestFloat:  999.99,
		TestTime:   time.Date(2017, 11, 17, 20, 34, 58, 651387237, time.UTC),
	}
}

func getTestRows(models []test.TestsModel) *sqlmock.Rows {
	var userFieldNames = []string{"id", "test_int", "test_string", "test_bool", "test_float", "test_time"}
	rows := sqlmock.NewRows(userFieldNames)
	for _, w := range models {
		rows = rows.AddRow(w.ID, w.TestInt, w.TestString, w.TestBool, w.TestFloat, w.TestTime)
	}
	return rows
}

func getTestModels(n int) []test.TestsModel {
	var ret []test.TestsModel
	for i := 0; i < n; i++ {
		u := test.TestsModel{
			ID:         999 + int64(i),
			TestInt:    1234,
			TestString: "Test my string",
			TestBool:   true,
			TestFloat:  999.99,
			TestTime:   time.Date(2017, 11, 17, 20, 34, 58, 651387237, time.UTC),
		}
		ret = append(ret, u)
	}

	return ret
}

func newDB() (sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("can't create sqlmock: %s", err)
	}

	d := postgres.New(
		postgres.Config{
			Conn:       db,
			DriverName: "postgresql",
		},
	)

	gormDB, gerr := gorm.Open(d, &gorm.Config{})
	if gerr != nil {
		log.Fatalf("can't open gorm connection: %s", err)
	}

	return mock, gormDB.Set("gorm:update_column", true)
}

func checkMock(t *testing.T, mock sqlmock.Sqlmock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestNewDatabase(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)

	dbTest := NewDatabase(db)
	conf, _ := dbTest.Orm.DB()

	err := conf.Ping()

	assert.Nil(t, err)
}

func TestDatabase_FetchOne(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)

	expectedModel := getTestModels(1)

	rows := getTestRows(expectedModel)
	req := `SELECT * FROM "tests" WHERE "tests"."id" = $1 AND "tests"."id" = $2 ORDER BY "tests"."id" LIMIT 1`

	m.ExpectQuery(regexp.QuoteMeta(req)).WillReturnRows(rows)

	dbTest := NewDatabase(db)

	model := test.TestsModel{ID: 999}

	resultTest := dbTest.FetchOne(&model)

	assert.True(t, true, resultTest)
	assert.Equal(t, expectedModel[0], model)
}

func TestDatabase_FetchAll(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)
	c := rand.Intn(50) //nolint:gosec

	expectedModel := getTestModels(c)

	rows := getTestRows(expectedModel)
	req := `SELECT * FROM "tests" WHERE "tests"."test_string" = $1`

	m.ExpectQuery(regexp.QuoteMeta(req)).WillReturnRows(rows)

	dbTest := NewDatabase(db)

	model := &test.TestsModel{TestString: "Test my string"}
	var foundModels []test.TestsModel

	resultTest := dbTest.FetchAll(model, &foundModels)
	assert.Equal(t, c, len(foundModels))
	assert.True(t, true, resultTest)
	assert.Equal(t, expectedModel, foundModels)
}

func TestDatabase_CountAll(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)
	var c int64
	var count = []string{"count"}
	c = rand.Int63n(50) //nolint:gosec
	countR := sqlmock.NewRows(count)
	countR = countR.AddRow(c)

	req := `SELECT count(*) FROM "tests" WHERE "tests"."test_string" = $1`

	m.ExpectQuery(regexp.QuoteMeta(req)).WillReturnRows(countR)

	dbTest := NewDatabase(db)
	model := test.TestsModel{TestString: "Test my string"}

	resultTest := dbTest.CountAll(&model)
	assert.Equal(t, c, resultTest)
}

func TestDatabase_Save(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)

	expectedModel := getTestModels(1)

	mutateModel := getTestModels(1)
	mutateModel[0].TestString = "Test my string 1234"
	mutatedRows := getTestRows(mutateModel)

	req := `UPDATE "tests" SET "test_string"=$1 WHERE "id" = $2`
	m.ExpectBegin()
	m.ExpectExec(regexp.QuoteMeta(req)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	m.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "tests" WHERE "id" = $1 AND "tests"."id" = $2`),
	).WithArgs(
		driver.Value(expectedModel[0].ID),
		driver.Value(expectedModel[0].ID),
	).WillReturnRows(mutatedRows)

	dbTest := NewDatabase(db)

	model := test.TestsModel{ID: 999, TestString: "Test my string 1234"}
	resultTest := dbTest.Save(&model)

	assert.NotEqual(t, expectedModel[0].TestString, model.TestString)
	assert.Nil(t, resultTest)
}

func TestDatabase_Create(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)

	expectedModel := getTestModels(1)
	expectedRows := getTestRows(expectedModel)
	model := getTestModelSingle()

	req := `INSERT INTO "tests" ("test_int","test_string","test_bool","test_float","test_time","id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta(req)).
		WithArgs(
			driver.Value(model.TestInt),
			driver.Value(model.TestString),
			driver.Value(model.TestBool),
			driver.Value(model.TestFloat),
			driver.Value(model.TestTime),
			driver.Value(model.ID),
		).WillReturnRows(expectedRows)
	m.ExpectCommit()

	dbTest := NewDatabase(db)

	resultTest := dbTest.Create(model)

	assert.Equal(t, &expectedModel[0], model)
	assert.Nil(t, resultTest)
}

func TestDatabase_Delete(t *testing.T) {
	m, db := newDB()
	defer checkMock(t, m)

	req := `DELETE FROM "tests" WHERE "tests"."test_string" = $1`
	m.ExpectBegin()
	m.ExpectExec(regexp.QuoteMeta(req)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()

	dbTest := NewDatabase(db)

	workerTest := test.TestsModel{TestString: "Test my string"}

	resultTest := dbTest.Delete(&workerTest)

	assert.Nil(t, resultTest)
}
