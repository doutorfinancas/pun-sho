package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/str"
)

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

type ShortyService struct {
	hostName               string
	logo                   string
	log                    *zap.Logger
	shortyRepository       *entity.ShortyRepository
	shortyAccessRepository *entity.ShortyAccessRepository
	qrSvc                  *QRCodeService
	publicIDLength         int
}

func NewShortyService(
	log *zap.Logger,
	shortyRepository *entity.ShortyRepository,
	shortyAccessRepository *entity.ShortyAccessRepository,
	qrSvc *QRCodeService,
	hostName,
	logo string,
	publicIDLength int,
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
	}
}

func (s *ShortyService) Create(req *request.CreateShorty) (*entity.Shorty, error) {
	m := &entity.Shorty{
		PublicID:         str.RandStringRunes(s.publicIDLength),
		Link:             req.Link,
		TTL:              req.TTL,
		RedirectionLimit: req.RedirectionLimit,
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

	if sh.TTL != nil && sh.TTL.Before(time.Now()) {
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
	if ua.Bot {
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

	m.Meta = m.ConvertMeta(req.Meta)

	if err := s.shortyAccessRepository.Create(m); err != nil {
		return nil, err
	}

	if status != StatusRedirected {
		return nil, errors.New(fmt.Sprintf("could not redirect due to status %s", status))
	}

	return sh, nil
}

func (s *ShortyService) List(limit, offset int) ([]*entity.Shorty, error) {
	shorties, err := s.shortyRepository.List(limit, offset)
	if err != nil {
		return []*entity.Shorty{}, err
	}

	return shorties, nil
}

func (s *ShortyService) FindShortyByID(id uuid.UUID, from, until string) (*entity.Shorty, error) {
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

		m.ShortyAccesses = sh

		m.Visits = len(sh)
		m.RedirectCount = CountRedirects(sh)

		return m, nil
	}

	sh := s.FindAllAccessesByShortyID(id)

	m.ShortyAccesses = sh

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
