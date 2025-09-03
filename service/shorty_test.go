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
	req := `INSERT INTO "shorties" ("public_id","link","ttl","redirection_limit","created_at","deleted_at","qr_code") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
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
					Link:             "www.cenas.pt",
					TTL:              nil,
					RedirectionLimit: nil,
				},
			},
			want: &entity.Shorty{
				ID:               uuid.UUID{},
				PublicID:         "",
				Link:             "",
				TTL:              nil,
				RedirectionLimit: nil,
				CreatedAt:        nil,
				DeletedAt:        nil,
				ShortyAccesses:   nil,
				ShortLink:        "",
				Visits:           0,
				RedirectCount:    0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := &ShortyService{
					hostName:          tt.fields.hostName,
					shortyRepository:  tt.fields.shortyRepository,
					allowedSocialBots: []string{}, // Lista vazia para teste
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

func TestCountRedirects(t *testing.T) {
	type args struct {
		accesses []entity.ShortyAccess
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test shorty accesses 'redirected' status count when equal to 0",
			args: args{
				accesses: []entity.ShortyAccess{
					{
						Status: StatusDeleted,
					},
					{
						Status: StatusExpired,
					},
					{
						Status: StatusBlocked,
					},
					{
						Status: StatusLimitReached,
					},
				},
			},
			want: 0,
		},
		{
			name: "Test shorty accesses 'redirected' status count",
			args: args{
				accesses: []entity.ShortyAccess{
					{
						Status: StatusDeleted,
					},
					{
						Status: StatusExpired,
					},
					{
						Status: StatusRedirected,
					},
					{
						Status: StatusBlocked,
					},
					{
						Status: StatusRedirected,
					},
					{
						Status: StatusLimitReached,
					},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CountRedirects(tt.args.accesses), "CountRedirects(%v)", tt.args.accesses)
		})
	}
}

func TestShortyService_IsSocialMediaBot(t *testing.T) {
	// Criar um ShortyService com a lista padrão de bots
	service := &ShortyService{
		allowedSocialBots: []string{
			"facebookexternalhit",
			"facebot",
			"googlebot",
			"linkedinbot",
			"twitterbot",
			"instagram",
			"instagrambot",
			"whatsapp",
			"slackbot",
			"telegrambot",
			"discordbot",
			"pinterestbot",
			"redditbot",
			"skypeuri",
			"applebot",
			"bingbot",
			"yandexbot",
		},
	}

	tests := []struct {
		name      string
		userAgent string
		want      bool
	}{
		{
			name:      "Facebook External Hit",
			userAgent: "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			want:      true,
		},
		{
			name:      "Google Bot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			want:      true,
		},
		{
			name:      "LinkedIn Bot",
			userAgent: "LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)",
			want:      true,
		},
		{
			name:      "Twitter Bot",
			userAgent: "Twitterbot/1.0",
			want:      true,
		},
		{
			name:      "Regular Chrome Browser",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			want:      false,
		},
		{
			name:      "Regular Safari Browser",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			want:      false,
		},
		{
			name:      "Malicious Bot",
			userAgent: "BadBot/1.0",
			want:      false,
		},
		{
			name:      "WhatsApp Bot",
			userAgent: "WhatsApp/2.19.81 A",
			want:      true,
		},
		{
			name:      "Instagram Bot",
			userAgent: "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)",
			want:      true,
		},
		{
			name:      "Instagram Bot 2",
			userAgent: "instagrambot",
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, service.isSocialMediaBot(tt.userAgent))
		})
	}
}
