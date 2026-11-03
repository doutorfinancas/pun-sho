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

func TestLabelsArrayLiteral(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		want   string
	}{
		{name: "plain", labels: []string{"clinica", "promo"}, want: `{"clinica","promo"}`},
		{name: "NULL stays a string", labels: []string{"NULL"}, want: `{"NULL"}`},
		{name: "comma inside a label", labels: []string{"a,b"}, want: `{"a,b"}`},
		{name: "quote and backslash escaped", labels: []string{`q"x`, `b\s`}, want: `{"q\"x","b\\s"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, labelsArrayLiteral(tt.labels))
		})
	}
}
