package entity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/doutorfinancas/pun-sho/test"
)

func TestShortyRepository_ListWithAccessData_CancelledContext(t *testing.T) {
	m, db := test.NewMockDB()
	defer test.CheckMockDB(t, m)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := ShortyRepository{Repository: test.NewMockRepository(db)}

	_, err := repo.ListWithAccessData(ctx, false, []string{"clinica"}, "", nil, nil, 1000, 0)

	assert.ErrorIs(t, err, context.Canceled)
}
