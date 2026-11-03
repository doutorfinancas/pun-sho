package api

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/service"
	"github.com/doutorfinancas/pun-sho/test"
)

func TestShortenerHandler_ListVisitCountsSince(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.MustParse("8e0ee302-a208-425e-8a8b-2442b6f35873")
	since := time.Now().UTC().Add(-26 * time.Hour).Truncate(time.Second).Format(time.RFC3339)

	tests := []struct {
		name       string
		query      string
		dbRows     [][]driver.Value
		dbErr      error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "returns visits and redirects per link",
			query:      "since=" + url.QueryEscape(since) + "&labels=clinica",
			dbRows:     [][]driver.Value{{id, int64(12), int64(9)}},
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":"8e0ee302-a208-425e-8a8b-2442b6f35873","visits":12,"redirects":9}]`,
		},
		{
			name:       "no visits returns an empty list",
			query:      "since=" + url.QueryEscape(since),
			dbRows:     [][]driver.Value{},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "missing since",
			query:      "labels=clinica",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status":"error","message":["since is required"]}`,
		},
		{
			name:       "since not RFC3339",
			query:      "since=2026-11-02",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status":"error","message":["since must be an RFC3339 timestamp"]}`,
		},
		{
			name:       "since beyond the window",
			query:      "since=2020-01-01T00:00:00Z",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status":"error","message":["since must be within the last 90 days"]}`,
		},
		{
			name:       "database failure",
			query:      "since=" + url.QueryEscape(since),
			dbErr:      errors.New("db down"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"status":"error","message":["could not load visit counts"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, db := test.NewMockDB()
			defer test.CheckMockDB(t, mock)

			if tt.dbRows != nil || tt.dbErr != nil {
				q := mock.ExpectQuery(regexp.QuoteMeta(`FROM recent r`))
				if tt.dbErr != nil {
					q.WillReturnError(tt.dbErr)
				} else {
					q.WillReturnRows(test.GenerateMockRows([]string{"id", "visits", "redirects"}, tt.dbRows))
				}
			}

			accessRepo := &entity.ShortyAccessRepository{Repository: test.NewMockRepository(db)}
			svc := service.NewShortyService(nil, nil, accessRepo, nil, "", "", 0, nil)
			h := NewShortenerHandler(svc, nil)

			r := gin.New()
			h.Routes(r.Group("/api/v1/short"))

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/short/visits?"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
