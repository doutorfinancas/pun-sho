package service

import (
	"database/sql/driver"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/test"
)

const browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"

// pointerValuerConverter lets sqlmock encode entity.Meta, whose Value method has a pointer receiver.
type pointerValuerConverter struct{}

func (pointerValuerConverter) ConvertValue(v interface{}) (driver.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() != reflect.Ptr {
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		if valuer, ok := ptr.Interface().(driver.Valuer); ok {
			return valuer.Value()
		}
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}

func TestShortyService_CreateVisit(t *testing.T) {
	shortyID := uuid.MustParse("8e0ee302-a208-425e-8a8b-2442b6f35873")
	limit := func(n int) *int { return &n }

	tests := []struct {
		name             string
		redirectionLimit *int
		countRows        *sqlmock.Rows
		countErr         error
		wantStatus       string
		wantInsert       bool
		wantErr          bool
	}{
		{
			name:       "no limit: redirects without counting accesses",
			wantStatus: StatusRedirected,
			wantInsert: true,
		},
		{
			name:             "zero limit means unlimited: no counting",
			redirectionLimit: limit(0),
			wantStatus:       StatusRedirected,
			wantInsert:       true,
		},
		{
			name:             "under limit: redirects",
			redirectionLimit: limit(5),
			countRows:        test.GenerateMockRows([]string{"count"}, [][]driver.Value{{int64(4)}}),
			wantStatus:       StatusRedirected,
			wantInsert:       true,
		},
		{
			name:             "limit reached: records the visit and refuses",
			redirectionLimit: limit(5),
			countRows:        test.GenerateMockRows([]string{"count"}, [][]driver.Value{{int64(5)}}),
			wantStatus:       StatusLimitReached,
			wantInsert:       true,
			wantErr:          true,
		},
		{
			name:             "count failure is returned, not treated as zero",
			redirectionLimit: limit(5),
			countErr:         errors.New("db down"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, db := test.NewMockDBWithConverter(pointerValuerConverter{})
			defer test.CheckMockDB(t, mock)

			shortyCols := []string{"id", "public_id", "link", "redirection_limit"}
			var limitVal driver.Value
			if tt.redirectionLimit != nil {
				limitVal = int64(*tt.redirectionLimit)
			}
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shorties"`)).
				WillReturnRows(test.GenerateMockRows(shortyCols, [][]driver.Value{{shortyID, "f2d9xGt", "https://example.pt", limitVal}}))

			if tt.countRows != nil || tt.countErr != nil {
				q := mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "shorty_accesses"`)).
					WithArgs(shortyID, StatusRedirected)
				if tt.countErr != nil {
					q.WillReturnError(tt.countErr)
				} else {
					q.WillReturnRows(tt.countRows)
				}
			}

			if tt.wantInsert {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shorty_accesses"`)).
					WithArgs(
						sqlmock.AnyArg(), shortyID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tt.wantStatus,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnRows(test.GenerateMockRows([]string{"id"}, [][]driver.Value{{uuid.New()}}))
				mock.ExpectCommit()
			}

			var tablesRead []string
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:tables_read", func(tx *gorm.DB) {
				tablesRead = append(tablesRead, tx.Statement.Table)
			}))

			fresh := func() *gorm.DB { return db.Session(&gorm.Session{NewDB: true}) }
			s := &ShortyService{
				shortyRepository:       &entity.ShortyRepository{Repository: test.NewMockRepository(fresh())},
				shortyAccessRepository: &entity.ShortyAccessRepository{Repository: test.NewMockRepository(fresh())},
				allowedSocialBots:      []string{},
			}

			got, err := s.CreateVisit("f2d9xGt", &request.Redirect{UserAgent: browserUA, IP: "1.2.3.4"})

			if tt.countRows == nil && tt.countErr == nil {
				assert.Equal(t, []string{"shorties"}, tablesRead, "unlimited links must not read shorty_accesses")
			}

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "https://example.pt", got.Link)
		})
	}
}
