package service

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/test"
)

func TestShortyService_VisitCountsSince(t *testing.T) {
	id := uuid.MustParse("8e0ee302-a208-425e-8a8b-2442b6f35873")

	tests := []struct {
		name      string
		since     time.Time
		wantQuery bool
		wantErr   string
	}{
		{name: "one day ago", since: time.Now().Add(-26 * time.Hour), wantQuery: true},
		{name: "at the window edge", since: time.Now().Add(-MaxVisitCountsWindow + time.Minute), wantQuery: true},
		{name: "beyond the window", since: time.Now().Add(-MaxVisitCountsWindow - time.Hour), wantErr: "since must be within the last 90 days"},
		{name: "in the future", since: time.Now().Add(time.Hour), wantErr: "since must not be in the future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, db := test.NewMockDB()
			defer test.CheckMockDB(t, mock)

			if tt.wantQuery {
				mock.ExpectQuery(regexp.QuoteMeta(`FROM recent r`)).
					WillReturnRows(test.GenerateMockRows(
						[]string{"id", "visits", "redirects"},
						[][]driver.Value{{id, int64(3), int64(2)}},
					))
			}

			s := &ShortyService{shortyAccessRepository: &entity.ShortyAccessRepository{Repository: test.NewMockRepository(db)}}

			got, err := s.VisitCountsSince(context.Background(), []string{"clinica"}, tt.since)

			if tt.wantErr != "" {
				var valErr *ValidationError
				require.ErrorAs(t, err, &valErr)
				assert.Equal(t, tt.wantErr, valErr.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []entity.ShortyVisitCount{{ID: id, Visits: 3, Redirects: 2}}, got)
		})
	}
}
