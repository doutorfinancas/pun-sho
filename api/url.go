package api

import (
	"fmt"
	"net/http"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/service"
	"github.com/gin-gonic/gin"

	"github.com/doutorfinancas/pun-sho/str"
)

type urlHandler struct {
	unknownPage string
	service     *service.ShortyService
}

func NewURLHandler(unknownPage string, svc *service.ShortyService) HTTPHandler {
	return &urlHandler{
		unknownPage: unknownPage,
		service:     svc,
	}
}

func (h *urlHandler) Routes(rg *gin.RouterGroup) {
	rg.GET("/:slug", h.RedirectLinkIfExists)
	rg.GET("/:slug/", h.RedirectLinkIfExists)
}

func (h *urlHandler) Group() *string {
	return str.ToStringNil("s")
}

func (h *urlHandler) RedirectLinkIfExists(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.Redirect(http.StatusMovedPermanently, h.unknownPage)
		return
	}

	meta := c.Request.URL.Query()
	for k, v := range c.Request.Header {
		// by default, if same key is sent, we override with header information since it's harder
		// to mess with it in user land (browser will be on our side)
		meta[k] = v
	}

	req := &request.Redirect{
		UserAgent: c.Request.UserAgent(),
		IP:        c.ClientIP(),
		Meta:      meta,
		Extra:     fmt.Sprintf("Map: %v", c.Request.URL.Query()),
	}

	sho, err := h.service.CreateVisit(slug, req)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, h.unknownPage)
		return
	}

	c.Header("Cache-Control", "private, max-age=90")
	c.Redirect(http.StatusMovedPermanently, sho.Link)
}
