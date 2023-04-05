package service

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/test"
)

func TestShortyService_Create(t *testing.T) {
	type fields struct {
		hostName         string
		shortyRepository *entity.ShortyRepository
	}
	type args struct {
		req *request.CreateShorty
	}

	m, db := test.NewMockDB()
	rows := test.GenerateMockRows(
		[]string{
			"created_at",
		},
		[][]driver.Value{
			{
				time.Now(),
			},
		},
	)
	req := `INSERT INTO "shorties" ("public_id","link","ttl","created_at","deleted_at") VALUES ($1,$2,$3,$4,$5)`
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta(req)).WithArgs().WillReturnRows(rows)
	m.ExpectCommit()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Shorty
		wantErr bool
	}{
		{
			name: "teste1",
			fields: fields{
				hostName: "batatas.pt",
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				req: &request.CreateShorty{
					Link: "www.cenas.pt",
					TTL:  nil,
				},
			},
			want: &entity.Shorty{
				ID:             uuid.UUID{},
				PublicID:       "",
				Link:           "",
				TTL:            nil,
				CreatedAt:      nil,
				DeletedAt:      nil,
				ShortyAccesses: nil,
				ShortLink:      "",
				Visits:         0,
				RedirectCount:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &ShortyService{
					hostName:         tt.fields.hostName,
					shortyRepository: tt.fields.shortyRepository,
				}
				got, err := s.Create(tt.args.req)
				if (err != nil) != tt.wantErr {
					t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				assert.Equal(t, "batatas.pt/s/"+got.PublicID, got.ShortLink)
			},
		)
	}
}
