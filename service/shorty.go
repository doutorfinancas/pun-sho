package service

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/str"
)

var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

const (
	DefaultPublicIDLength = 10
	StatusRedirected      = "redirected"
	StatusBlocked         = "blocked"
	StatusExpired         = "expired"
	StatusLimitReached    = "limit_reached"
	StatusDeleted         = "deleted"
	VersionStringify      = "%s %s"
	TransparentBackground = "transparent"
)

// isSocialMediaBot checks if the User-Agent corresponds to an allowed social media bot
// Uses precise matching to prevent false positives from malicious user agents
func (s *ShortyService) isSocialMediaBot(userAgent string) bool {
	userAgentLower := strings.ToLower(userAgent)

	for _, bot := range s.allowedSocialBots {
		botLower := strings.ToLower(bot)

		// Check for exact match or bot name followed by common separators
		// This prevents false positives like "malicious-facebookexternalhit-exploit"
		if userAgentLower == botLower ||
			strings.HasPrefix(userAgentLower, botLower+"/") ||
			strings.HasPrefix(userAgentLower, botLower+" ") ||
			strings.Contains(userAgentLower, " "+botLower+"/") ||
			strings.Contains(userAgentLower, " "+botLower+" ") {
			return true
		}
	}
	return false
}

type ShortyService struct {
	hostName               string
	logo                   string
	log                    *zap.Logger
	shortyRepository       *entity.ShortyRepository
	shortyAccessRepository *entity.ShortyAccessRepository
	qrSvc                  *QRCodeService
	geoSvc                 *GeoIPService
	publicIDLength         int
	allowedSocialBots      []string
}

func NewShortyService(
	log *zap.Logger,
	shortyRepository *entity.ShortyRepository,
	shortyAccessRepository *entity.ShortyAccessRepository,
	qrSvc *QRCodeService,
	hostName,
	logo string,
	publicIDLength int,
	allowedSocialBots []string,
) *ShortyService {
	if publicIDLength == 0 {
		publicIDLength = DefaultPublicIDLength
	}

	return &ShortyService{
		hostName:               strings.TrimSuffix(hostName, "/"),
		logo:                   logo,
		log:                    log,
		shortyRepository:       shortyRepository,
		shortyAccessRepository: shortyAccessRepository,
		qrSvc:                  qrSvc,
		publicIDLength:         publicIDLength,
		allowedSocialBots:      allowedSocialBots,
	}
}

func (s *ShortyService) SetGeoIPService(geoSvc *GeoIPService) {
	s.geoSvc = geoSvc
}

func (s *ShortyService) GetHostName() string {
	return s.hostName
}

func (s *ShortyService) Create(req *request.CreateShorty) (*entity.Shorty, error) {
	publicID := str.RandStringRunes(s.publicIDLength)

	if req.Slug != nil && *req.Slug != "" {
		slug := *req.Slug
		if len(slug) < 3 {
			return nil, fmt.Errorf("slug must be at least 3 characters")
		}
		if !slugRegex.MatchString(slug) {
			return nil, fmt.Errorf("slug contains invalid characters")
		}
		exists, err := s.shortyRepository.ExistsByPublicID(slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("slug already in use")
		}
		publicID = slug
	}

	link := req.Link
	if req.UTM != nil && !req.UTM.IsEmpty() {
		link = appendUTMParams(link, req.UTM)
	}

	m := &entity.Shorty{
		PublicID:         publicID,
		Link:             link,
		TTL:              req.TTL,
		RedirectionLimit: req.RedirectionLimit,
		Labels:           req.Labels,
	}

	m.ShortLink = fmt.Sprintf("%s/s/%s", s.hostName, m.PublicID)
	if req.QRCode != nil && req.QRCode.Create {
		q, err := s.qrSvc.Generate(req.QRCode, m.ShortLink)
		if err != nil {
			return nil, err
		}

		m.QRCode = q
	}

	if err := s.shortyRepository.Create(m); err != nil {
		return nil, err
	}

	return m, nil
}

func appendUTMParams(link string, utm *request.UTMParams) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	q := u.Query()
	if utm.Source != "" {
		q.Set("utm_source", utm.Source)
	}
	if utm.Medium != "" {
		q.Set("utm_medium", utm.Medium)
	}
	if utm.Campaign != "" {
		q.Set("utm_campaign", utm.Campaign)
	}
	if utm.Term != "" {
		q.Set("utm_term", utm.Term)
	}
	if utm.Content != "" {
		q.Set("utm_content", utm.Content)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func (s *ShortyService) RegenerateQR(qrReq *request.QRCode, shortLink string) (string, error) {
	return s.qrSvc.Generate(qrReq, shortLink)
}

func (s *ShortyService) Update(req *request.UpdateShorty, m *entity.Shorty) (*entity.Shorty, error) {
	if req.Link != "" {
		m.Link = req.Link
	}

	if req.Cancel {
		now := time.Now()
		m.DeletedAt = &now
	}

	if req.TTL != nil {
		m.TTL = req.TTL
	}

	if req.RedirectionLimit != nil {
		m.RedirectionLimit = req.RedirectionLimit
	}

	if req.Labels != nil {
		m.Labels = req.Labels
	}

	if err := s.shortyRepository.Save(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *ShortyService) CreateVisit(publicID string, req *request.Redirect) (*entity.Shorty, error) {
	sh, err := s.FindShortyByPublicID(publicID)
	if err != nil {
		return nil, err
	}

	status := StatusRedirected
	if sh.DeletedAt != nil {
		status = StatusDeleted
	}

	if sh.TTL != nil && sh.TTL.Year() > 1 && sh.TTL.Before(time.Now()) {
		status = StatusExpired
	}

	accessList := s.FindAllAccessesByShortyID(sh.ID)
	redirectCount := CountRedirects(accessList)
	if sh.RedirectionLimit != nil && *sh.RedirectionLimit != 0 {
		if redirectCount >= *sh.RedirectionLimit {
			status = StatusLimitReached
		}
	}

	ua := useragent.Parse(req.UserAgent)
	if ua.Bot && !s.isSocialMediaBot(req.UserAgent) {
		status = StatusBlocked
	}

	m := &entity.ShortyAccess{
		ShortyID:        sh.ID,
		Status:          status,
		UserAgent:       req.UserAgent,
		IPAddress:       req.IP,
		Browser:         fmt.Sprintf(VersionStringify, ua.Name, ua.Version),
		OperatingSystem: fmt.Sprintf(VersionStringify, ua.OS, ua.OSVersion),
		Extra:           req.Extra,
	}

	if s.geoSvc != nil {
		m.Country, m.City = s.geoSvc.Lookup(req.IP)
	}

	m.Meta = m.ConvertMeta(req.Meta)

	if err := s.shortyAccessRepository.Create(m); err != nil {
		return nil, err
	}

	if status != StatusRedirected {
		return nil, fmt.Errorf("could not redirect due to status %s", status)
	}

	return sh, nil
}

func (s *ShortyService) List(withQR bool, labels []string, limit, offset int) ([]*entity.ShortyForList, error) {
	shorties, err := s.shortyRepository.ListWithAccessData(withQR, labels, limit, offset)
	if err != nil {
		return nil, err
	}

	return shorties, nil
}

func (s *ShortyService) FindShortyByID(id uuid.UUID, from, until string, showAccesses bool) (*entity.Shorty, error) {
	m := &entity.Shorty{
		ID: id,
	}

	if err := s.shortyRepository.Database.FetchOne(m); err != nil {
		return nil, err
	}

	if from != "" && until != "" {
		fromTime, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return nil, err
		}

		untilTime, err := time.Parse(time.DateTime, until+" 23:59:59")
		if err != nil {
			return nil, err
		}

		sh := s.FindAllAccessesByShortyIDAndDateRange(id, &fromTime, &untilTime)

		if showAccesses {
			m.ShortyAccesses = sh
		}

		m.Visits = len(sh)
		m.RedirectCount = CountRedirects(sh)

		return m, nil
	}

	sh := s.FindAllAccessesByShortyID(id)

	if showAccesses {
		m.ShortyAccesses = sh
	}

	m.Visits = len(sh)
	m.RedirectCount = CountRedirects(sh)

	return m, nil
}

func (s *ShortyService) FindShortyByPublicID(publicID string) (*entity.Shorty, error) {
	m := &entity.Shorty{
		PublicID: publicID,
	}

	if err := s.shortyRepository.Database.FetchOne(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *ShortyService) DeleteShortyByUUID(id uuid.UUID) error {
	return s.shortyRepository.Delete(id)
}

func (s *ShortyService) FindAllAccessesByShortyID(id uuid.UUID) []entity.ShortyAccess {
	var sh []entity.ShortyAccess

	_ = s.shortyAccessRepository.Database.FetchAll(
		&entity.ShortyAccess{ShortyID: id},
		&sh,
	)

	return sh
}

func (s *ShortyService) FindAllAccessesByShortyIDAndDateRange(
	id uuid.UUID,
	from,
	until *time.Time,
) []entity.ShortyAccess {
	var sh []entity.ShortyAccess

	_ = s.shortyAccessRepository.Database.Orm.
		Model(&entity.ShortyAccess{}).
		Where("shorty_id = ?", id).
		Where("created_at BETWEEN ? AND ?", from, until).Scan(&sh)

	return sh
}

func CountRedirects(accesses []entity.ShortyAccess) int {
	var redirects int
	for _, access := range accesses {
		if access.Status == StatusRedirected {
			redirects++
		}
	}
	return redirects
}
