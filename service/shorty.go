package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doutorfinancas/pun-sho/buf"
	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/str"
)

const (
	PublicIDSize          = 10
	StatusRedirected      = "redirected"
	StatusBlocked         = "blocked"
	StatusExpired         = "expired"
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
}

func NewShortyService(
	hostName,
	logo string,
	log *zap.Logger,
	shortyRepository *entity.ShortyRepository,
	shortyAccessRepository *entity.ShortyAccessRepository,
) *ShortyService {
	return &ShortyService{
		hostName:               strings.TrimSuffix(hostName, "/"),
		logo:                   logo,
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

	m.ShortLink = fmt.Sprintf("%s/s/%s", s.hostName, m.PublicID)
	if req.QRCode != nil && req.QRCode.Create {
		qrc, err := qrcode.New(m.ShortLink)
		if err != nil {
			return nil, err
		}

		bgColor := standard.WithBgColorRGBHex("#ffffff")
		fgColor := standard.WithFgColorRGBHex("#000000")

		if str.SubString(req.QRCode.BgColor, 0, 1) == "#" {
			bgColor = standard.WithBgColorRGBHex(req.QRCode.BgColor)
		}

		if req.QRCode.BgColor == TransparentBackground {
			bgColor = standard.WithBgTransparent()
		}

		if str.SubString(req.QRCode.FgColor, 0, 1) == "#" {
			fgColor = standard.WithFgColorRGBHex(req.QRCode.FgColor)
		}

		options := []standard.ImageOption{
			bgColor,
			fgColor,
			standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		}

		if s.logo != "" {
			fmt.Println(s.logo)
			options = append(options, standard.WithLogoImageFilePNG(s.logo))
		}

		if req.QRCode.Width > 0 {
			options = append(options, standard.WithQRWidth(uint8(req.QRCode.Width)))
		}

		if req.QRCode.BorderWidth > 0 {
			options = append(options, standard.WithBorderWidth(req.QRCode.BorderWidth))
		}

		if req.QRCode.Shape == "circle" {
			options = append(options, standard.WithCircleShape())
		}

		var b []byte
		x := bytes.NewBuffer(b)
		w := buf.NewWriteCloser(x)
		wr := standard.NewWithWriter(w, options...)

		err = qrc.Save(wr)
		if err != nil {
			return nil, err
		}

		m.QRCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(x.Bytes())
	}

	if err := s.shortyRepository.Create(m); err != nil {
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
