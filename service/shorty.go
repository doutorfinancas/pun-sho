package service

import (
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/str"
)

const PublicIDSize = 10

type ShortyService struct {
	log                    *zap.Logger
	shortyRepository       *entity.ShortyRepository
	shortyAccessRepository *entity.ShortyAccessRepository
}

func NewShortyService(
	log *zap.Logger,
	shortyRepository *entity.ShortyRepository,
	shortyAccessRepository *entity.ShortyAccessRepository,
) *ShortyService {
	return &ShortyService{
		log:                    log,
		shortyRepository:       shortyRepository,
		shortyAccessRepository: shortyAccessRepository,
	}
}

func (s ShortyService) Create(req *request.CreateShorty) (*entity.Shorty, error) {
	m := &entity.Shorty{
		PublicID: str.RandStringRunes(PublicIDSize),
		Link:     req.Link,
		TTL:      req.TTL,
	}

	if err := s.shortyRepository.Create(m); err != nil {
		return nil, err
	}

	m.ShortLink = fmt.Sprintf("%s/s/%s", "http://localhost:8080", m.PublicID)

	return m, nil
}

func (s ShortyService) List(limit, offset int) ([]*entity.Shorty, error) {
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

	return m, nil
}
