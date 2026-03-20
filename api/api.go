package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/service"
)

type API struct {
	BaseGinServer
	log          *zap.Logger
	config       *Config
	shortySvc    *service.ShortyService
	qrSvc        *service.QRCodeService
	authSvc      *service.AuthService
	analyticsSvc *service.AnalyticsService
}

func NewAPI(
	log *zap.Logger,
	config *Config,
	shortyService *service.ShortyService,
	qrSvc *service.QRCodeService,
	authSvc *service.AuthService,
	analyticsSvc *service.AnalyticsService,
) *API {
	return &API{
		log:          log,
		config:       config,
		shortySvc:    shortyService,
		qrSvc:        qrSvc,
		authSvc:      authSvc,
		analyticsSvc: analyticsSvc,
	}
}

func (a *API) Run() {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()

	g.Use(
		gin.Recovery(),
		ginzap.Ginzap(a.log, time.RFC3339, true),
		gzip.Gzip(gzip.DefaultCompression),
	)

	// Load templates
	LoadTemplates(a.log)

	// Static file serving
	g.Static("/static", "./static")

	// Public redirect routes
	a.PushHandlerWithGroup(NewURLHandler(a.config.UnknownPage, a.shortySvc), g.Group("/"))

	// API routes (token auth)
	authMiddleware := NewAuthenticationMiddleware(a.config.Token)
	apiGroup := g.Group("/api/v1")
	apiGroup.Use(authMiddleware.Authenticated)
	a.PushHandlerWithGroup(NewShortenerHandler(a.shortySvc), apiGroup)
	a.PushHandlerWithGroup(NewPreviewHandler(a.qrSvc), apiGroup)

	// App routes (session auth)
	appGroup := g.Group("/app")
	sessionMiddleware := NewSessionMiddleware(a.authSvc)

	// Auth handler (login/logout — no session required)
	authHandler := NewAuthHandler(a.log, a.authSvc, a.config.CookieDomain, a.config.DisableLocalLogin, int(a.config.GetSessionDuration().Seconds()))
	if a.config.MicrosoftTenantID != "" {
		authHandler.ConfigureMicrosoftOAuth(
			a.config.MicrosoftTenantID,
			a.config.MicrosoftClientID,
			a.config.MicrosoftSecret,
			a.config.GetMicrosoftAllowedGroups(),
		)
	}
	authHandler.Validate()
	a.PushHandlerWithGroup(authHandler, appGroup)

	// Protected app routes (session required)
	protectedGroup := appGroup.Group("")
	protectedGroup.Use(sessionMiddleware.RequireSession)
	frontendHandler := NewFrontendHandler(
		a.log,
		a.shortySvc,
		a.analyticsSvc,
		a.authSvc,
		a.config.HostName,
	)
	a.PushHandlerWithGroup(frontendHandler, protectedGroup)

	a.log.Info("API Starting", zap.Int("port", a.config.Port))

	if err := g.Run(fmt.Sprintf(":%d", a.config.Port)); err != nil {
		a.log.Fatal(err.Error())
	}
}

func validateLimitAndOffset(limitStr, offsetStr string) (int, int, string, error) {
	limit := 0
	offset := 0
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, "invalid limit parameter", err
		}
	} else {
		limit = 0
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, "invalid offset parameter", err
		}
	} else {
		offset = 0
	}

	return limit, offset, "", err
}
