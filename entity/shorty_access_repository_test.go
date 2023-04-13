package entity

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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
