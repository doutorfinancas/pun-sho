package api

import (
	"fmt"
	"net/http"
	"strings"

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
		c.Redirect(http.StatusFound, h.unknownPage)
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
		IP:        ReadUserIP(c.Request),
		Meta:      meta,
		Extra:     fmt.Sprintf("Map: %v", c.Request.URL.Query()),
	}

	sho, err := h.service.CreateVisit(slug, req)
	if err != nil {
		c.Redirect(http.StatusFound, h.unknownPage)
		return
	}

	c.Header("Cache-Control", "private, max-age=90")
	c.Redirect(http.StatusFound, sho.Link)
}

func ReadUserIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
