package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/str"
	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"go.uber.org/zap"
)

const (
	PublicIDSize     = 10
	StatusRedirected = "redirected"
	StatusBlocked    = "blocked"
	StatusExpired    = "expired"
	StatusDeleted    = "deleted"
	VersionStringify = "%s %s"
)

type ShortyService struct {
	hostName               string
	log                    *zap.Logger
	shortyRepository       *entity.ShortyRepository
	shortyAccessRepository *entity.ShortyAccessRepository
}

func NewShortyService(
	hostName string,
	log *zap.Logger,
	shortyRepository *entity.ShortyRepository,
	shortyAccessRepository *entity.ShortyAccessRepository,
) *ShortyService {
	return &ShortyService{
		hostName:               strings.TrimSuffix(hostName, "/"),
		log:                    log,
		shortyRepository:       shortyRepository,
		shortyAccessRepository: shortyAccessRepository,
	}
}

func (s *ShortyService) Create(req *request.CreateShorty) (*entity.Shorty, error) {
	m := &entity.Shorty{
		PublicID: str.RandStringRunes(PublicIDSize),
		Link:     req.Link,
		TTL:      req.TTL,
	}

	if err := s.shortyRepository.Create(m); err != nil {
		return nil, err
	}

	m.ShortLink = fmt.Sprintf("%s/s/%s", s.hostName, m.PublicID)

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

func (s *ShortyService) FindShortyByID(id uuid.UUID) (*entity.Shorty, error) {
	m := &entity.Shorty{
		ID: id,
	}

	if err := s.shortyRepository.Database.FetchOne(m); err != nil {
		return nil, err
	}

	var sh []entity.ShortyAccess

	_ = s.shortyAccessRepository.Database.FetchAll(
		&entity.ShortyAccess{ShortyID: id},
		&sh,
	)

	spew.Dump(sh)

	m.ShortyAccesses = sh

	m.Visits = len(sh)
	m.RedirectCount = func(a []entity.ShortyAccess) (tmp int) {
		for _, v := range a {
			if v.Status == StatusRedirected {
				tmp++
			}
		}
		return
	}(sh)

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
