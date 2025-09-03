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
					allowedSocialBots: []string{}, // Empty list for testing
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
	testCases := []struct {
		name           string
		allowedBots    []string
		userAgent      string
		expectedResult bool
	}{
		// Test case 1: No bots allowed (empty configuration)
		{
			name:           "No bots allowed - Facebook bot should be blocked",
			allowedBots:    []string{},
			userAgent:      "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			expectedResult: false,
		},
		{
			name:           "No bots allowed - Google bot should be blocked",
			allowedBots:    []string{},
			userAgent:      "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expectedResult: false,
		},
		{
			name:           "No bots allowed - Regular browser should pass",
			allowedBots:    []string{},
			userAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			expectedResult: false,
		},

		// Test case 2: Only Facebook allowed
		{
			name:           "Only Facebook allowed - Facebook bot should pass",
			allowedBots:    []string{"facebookexternalhit", "facebot"},
			userAgent:      "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
			expectedResult: true,
		},
		{
			name:           "Only Facebook allowed - Google bot should be blocked",
			allowedBots:    []string{"facebookexternalhit", "facebot"},
			userAgent:      "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expectedResult: false,
		},

		// Test case 3: Multiple bots allowed
		{
			name:           "Multiple bots - Instagram should pass",
			allowedBots:    []string{"facebookexternalhit", "googlebot", "instagram", "instagrambot"},
			userAgent:      "Instagram 76.0.0.15.395 Android",
			expectedResult: true,
		},
		{
			name:           "Multiple bots - LinkedIn should be blocked",
			allowedBots:    []string{"facebookexternalhit", "googlebot", "instagram"},
			userAgent:      "LinkedInBot/1.0",
			expectedResult: false,
		},

		// Test case 4: Case sensitivity
		{
			name:           "Case insensitive - Mixed case should work",
			allowedBots:    []string{"googlebot"},
			userAgent:      "Mozilla/5.0 (compatible; GoogleBot/2.1; +http://www.google.com/bot.html)",
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &ShortyService{
				allowedSocialBots: tc.allowedBots,
			}
			result := service.isSocialMediaBot(tc.userAgent)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
