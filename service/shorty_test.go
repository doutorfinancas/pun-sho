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
	req := `INSERT INTO "shorties" ("public_id","link","ttl","redirection_limit","created_at","deleted_at","qr_code","labels") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`
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
			name: "create shorty without labels",
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
		{
			name: "create shorty with labels",
			fields: fields{
				hostName: "batatas.pt",
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				req: &request.CreateShorty{
					Link:             "www.example.pt",
					TTL:              nil,
					RedirectionLimit: nil,
					Labels:           []string{"marketing", "campaign", "2024"},
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
				if tt.args.req.Labels != nil {
					assert.Equal(t, entity.StringArray(tt.args.req.Labels), got.Labels)
				}
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
		{
			name:           "Case insensitive - Config with mixed case",
			allowedBots:    []string{"FacebookExternalHit"},
			userAgent:      "facebookexternalhit/1.1",
			expectedResult: true,
		},

		// Test case 5: Security - False positive prevention
		{
			name:           "Security - Malicious bot with prefix should be blocked",
			allowedBots:    []string{"facebookexternalhit"},
			userAgent:      "malicious-facebookexternalhit-exploit",
			expectedResult: false,
		},
		{
			name:           "Security - Malicious bot with suffix should be blocked",
			allowedBots:    []string{"googlebot"},
			userAgent:      "googlebot-malicious/1.0",
			expectedResult: false,
		},
		{
			name:           "Security - Bot name in middle should be blocked",
			allowedBots:    []string{"linkedinbot"},
			userAgent:      "evil-linkedinbot-scraper",
			expectedResult: false,
		},

		// Test case 6: Valid bot patterns that should work
		{
			name:           "Valid - Bot with version",
			allowedBots:    []string{"facebookexternalhit"},
			userAgent:      "facebookexternalhit/1.1",
			expectedResult: true,
		},
		{
			name:           "Valid - Bot with space",
			allowedBots:    []string{"googlebot"},
			userAgent:      "Mozilla/5.0 (compatible; googlebot/2.1; +http://www.google.com/bot.html)",
			expectedResult: true,
		},
		{
			name:           "Valid - Exact bot name",
			allowedBots:    []string{"instagrambot"},
			userAgent:      "instagrambot",
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

func TestShortyService_ListWithLabels(t *testing.T) {
	type fields struct {
		shortyRepository *entity.ShortyRepository
	}
	type args struct {
		withQR bool
		labels []string
		limit  int
		offset int
	}

	m, db := test.NewMockDB()
	
	// Mock data for testing
	mockRows := test.GenerateMockRows(
		[]string{
			"id", "created_at", "deleted_at", "public_id", "link", 
			"ttl", "redirection_limit", "labels", "visits", "redirects",
		},
		[][]driver.Value{
			{
				uuid.New(), time.Now(), nil, "test1", "https://example1.com",
				nil, nil, "{marketing,promotional}", 10, 8,
			},
			{
				uuid.New(), time.Now(), nil, "test2", "https://example2.com",
				nil, nil, "{tech,development}", 5, 4,
			},
		},
	)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "list without labels filter",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				withQR: false,
				labels: nil,
				limit:  10,
				offset: 0,
			},
			wantErr: false,
		},
		{
			name: "list with single label filter",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				withQR: false,
				labels: []string{"marketing"},
				limit:  10,
				offset: 0,
			},
			wantErr: false,
		},
		{
			name: "list with multiple labels filter",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				withQR: false,
				labels: []string{"marketing", "tech"},
				limit:  10,
				offset: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			m, db = test.NewMockDB()
			mockRows = test.GenerateMockRows(
				[]string{
					"id", "created_at", "deleted_at", "public_id", "link", 
					"ttl", "redirection_limit", "labels", "visits", "redirects",
				},
				[][]driver.Value{
					{
						uuid.New(), time.Now(), nil, "test1", "https://example1.com",
						nil, nil, "{marketing,promotional}", 10, 8,
					},
				},
			)
			
			// Expect the query with or without labels
			if len(tt.args.labels) > 0 {
				expectedQuery := `SELECT s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.ttl, s.redirection_limit, s.labels, count(sa.id) as visits, sum(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END) as redirects FROM shorties s 
    INNER JOIN shorty_accesses sa 
        ON s.id = sa.shorty_id WHERE s.labels && ? GROUP BY s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.qr_code, s.ttl, s.redirection_limit, s.labels LIMIT ? OFFSET ?`
				m.ExpectQuery(regexp.QuoteMeta(expectedQuery)).WithArgs("{marketing}", 10, 0).WillReturnRows(mockRows)
			} else {
				expectedQuery := `SELECT s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.ttl, s.redirection_limit, s.labels, count(sa.id) as visits, sum(CASE WHEN sa.status = 'redirected' THEN 1 ELSE 0 END) as redirects FROM shorties s 
    INNER JOIN shorty_accesses sa 
        ON s.id = sa.shorty_id GROUP BY s.id, s.created_at, s.deleted_at, s.public_id, s.link, s.qr_code, s.ttl, s.redirection_limit, s.labels LIMIT ? OFFSET ?`
				m.ExpectQuery(regexp.QuoteMeta(expectedQuery)).WithArgs(10, 0).WillReturnRows(mockRows)
			}

			s := &ShortyService{
				shortyRepository: tt.fields.shortyRepository,
			}
			
			got, err := s.List(tt.args.withQR, tt.args.labels, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			assert.NotNil(t, got)
			if len(got) > 0 {
				assert.Equal(t, "test1", got[0].PublicID)
			}
			
			test.CheckMockDB(t, m)
		})
	}
}

func TestShortyService_UpdateWithLabels(t *testing.T) {
	type fields struct {
		shortyRepository *entity.ShortyRepository
	}
	type args struct {
		req *request.UpdateShorty
		m   *entity.Shorty
	}

	m, db := test.NewMockDB()
	
	// Mock existing shorty
	existingShorty := &entity.Shorty{
		ID:       uuid.New(),
		PublicID: "test123",
		Link:     "https://old-link.com",
		Labels:   entity.StringArray{"old", "label"},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update shorty with new labels",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				req: &request.UpdateShorty{
					Labels: []string{"marketing", "campaign", "2024"},
				},
				m: existingShorty,
			},
			wantErr: false,
		},
		{
			name: "update shorty with empty labels",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				req: &request.UpdateShorty{
					Labels: []string{},
				},
				m: existingShorty,
			},
			wantErr: false,
		},
		{
			name: "update shorty without changing labels",
			fields: fields{
				shortyRepository: &entity.ShortyRepository{
					Repository: test.NewMockRepository(db),
				},
			},
			args: args{
				req: &request.UpdateShorty{
					Link: "https://new-link.com",
				},
				m: existingShorty,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			m, db = test.NewMockDB()
			
			// Expect update query
			m.ExpectBegin()
			m.ExpectCommit()

			s := &ShortyService{
				shortyRepository: tt.fields.shortyRepository,
			}
			
			// Copy the existing shorty to avoid modifying the original
			shortyCopy := *tt.args.m
			
			got, err := s.Update(tt.args.req, &shortyCopy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			assert.NotNil(t, got)
			
			// Check if labels were updated correctly
			if tt.args.req.Labels != nil {
				assert.Equal(t, entity.StringArray(tt.args.req.Labels), got.Labels)
			} else {
				// Labels should remain unchanged
				assert.Equal(t, tt.args.m.Labels, got.Labels)
			}
			
			test.CheckMockDB(t, m)
		})
	}
}
