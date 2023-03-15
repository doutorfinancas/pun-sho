package api

import (
	"fmt"
	"time"

	"github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type API struct {
	BaseGinServer
	log    *zap.Logger
	config *Config
}

func NewAPI(log *zap.Logger, config *Config) *API {
	return &API{
		log:    log,
		config: config,
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

	a.PushHandlerWithGroup(NewURLHandler(a.config.UnknownPage), g.Group("/"))

	authMiddleware := NewAuthenticationMiddleware(a.config.Token)

	apiGroup := g.Group("/api/v1")
	apiGroup.Use(authMiddleware.Authenticated)
	a.PushHandlerWithGroup(NewShortenerHandler(), apiGroup)

	if err := g.Run(fmt.Sprintf(":%d", a.config.Port)); err != nil {
		a.log.Fatal(err.Error())
	}
}
