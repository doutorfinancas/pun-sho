package entity

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doutorfinancas/pun-sho/test"
)

func TestDatabase_ListByShortyID(t *testing.T) {
	m, db := test.NewMockDB()
	defer test.CheckMockDB(t, m)

	expectedModel := []ShortyAccess{
		{
			ID:              uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			CreatedAt:       nil,
			ShortyID:        uuid.MustParse("cf07d8b6-89c5-4bb8-bc58-51cf831ea499"),
			UserAgent:       "",
			IPAddress:       "",
			Extra:           "",
			OperatingSystem: "MACOS",
			Browser:         "CHROME",
			Status:          "",
		},
	}

	rows := test.GenerateMockRows(
		[]string{
			"id",
			"created_at",
			"shorty_id",
			"user_agent",
			"ip_address",
			"extra",
			"operating_system",
			"browser",
		},
		[][]driver.Value{
			{
				expectedModel[0].ID,
				expectedModel[0].CreatedAt,
				expectedModel[0].ShortyID,
				expectedModel[0].UserAgent,
				expectedModel[0].IPAddress,
				expectedModel[0].Extra,
				expectedModel[0].OperatingSystem,
				expectedModel[0].Browser,
			},
		},
	)
	req := `SELECT shorty_accesses.* FROM "shorty_accesses" WHERE shorty_accesses.shorty_id = $1 LIMIT 1`

	m.ExpectQuery(regexp.QuoteMeta(req)).WillReturnRows(rows)

	dbTest := ShortyAccessRepository{
		Repository: test.NewMockRepository(db),
	}

	resultTest, err := dbTest.ListByShortyUUID(uuid.MustParse("cf07d8b6-89c5-4bb8-bc58-51cf831ea499"), 1, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedModel[0], *resultTest[0])
}

func TestShortyAccessRepository_CountRedirectsByShortyID(t *testing.T) {
	m, db := test.NewMockDB()
	defer test.CheckMockDB(t, m)

	id := uuid.MustParse("cf07d8b6-89c5-4bb8-bc58-51cf831ea499")
	req := `SELECT count(*) FROM "shorty_accesses" WHERE shorty_id = $1 AND status = $2`
	m.ExpectQuery(regexp.QuoteMeta(req)).
		WithArgs(id, "redirected").
		WillReturnRows(test.GenerateMockRows([]string{"count"}, [][]driver.Value{{int64(42)}}))

	repo := ShortyAccessRepository{Repository: test.NewMockRepository(db)}

	got, err := repo.CountRedirectsByShortyID(id)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), got)
}

func TestShortyAccessRepository_VisitCountsSince(t *testing.T) {
	since := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	id := uuid.MustParse("cf07d8b6-89c5-4bb8-bc58-51cf831ea499")

	tests := []struct {
		name     string
		labels   []string
		wantSQL  string
		wantArgs []driver.Value
	}{
		{
			name:     "without labels",
			wantSQL:  `WHERE created_at >= $1 )`,
			wantArgs: []driver.Value{since},
		},
		{
			name:     "with labels",
			labels:   []string{"clinica", "promo"},
			wantSQL:  `WHERE (SELECT s.labels FROM shorties s WHERE s.id = r.shorty_id) && $2 ORDER BY r.shorty_id`,
			wantArgs: []driver.Value{since, `{"clinica","promo"}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, db := test.NewMockDB()
			defer test.CheckMockDB(t, m)

			m.ExpectQuery(regexp.QuoteMeta(tt.wantSQL)).
				WithArgs(tt.wantArgs...).
				WillReturnRows(test.GenerateMockRows(
					[]string{"id", "visits", "redirects"},
					[][]driver.Value{{id, int64(12), int64(9)}},
				))

			repo := ShortyAccessRepository{Repository: test.NewMockRepository(db)}

			got, err := repo.VisitCountsSince(context.Background(), tt.labels, since)

			require.NoError(t, err)
			assert.Equal(t, []ShortyVisitCount{{ID: id, Visits: 12, Redirects: 9}}, got)
		})
	}
}

func TestShortyAccessRepository_VisitCountsSince_Empty(t *testing.T) {
	m, db := test.NewMockDB()
	defer test.CheckMockDB(t, m)

	m.ExpectQuery(regexp.QuoteMeta(`FROM recent r`)).
		WillReturnRows(test.GenerateMockRows([]string{"id", "visits", "redirects"}, nil))

	repo := ShortyAccessRepository{Repository: test.NewMockRepository(db)}

	got, err := repo.VisitCountsSince(context.Background(), nil, time.Now())

	require.NoError(t, err)
	assert.NotNil(t, got, "an empty result must serialise as [] not null")
	assert.Empty(t, got)
}

func TestShortyAccessRepository_VisitCountsSince_CancelledContext(t *testing.T) {
	m, db := test.NewMockDB()
	defer test.CheckMockDB(t, m)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := ShortyAccessRepository{Repository: test.NewMockRepository(db)}

	_, err := repo.VisitCountsSince(ctx, nil, time.Now())

	assert.ErrorIs(t, err, context.Canceled)
}
